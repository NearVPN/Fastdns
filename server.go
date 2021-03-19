package fastdns

import (
	"errors"
	"net"
	"os"
	"runtime"
	"time"
)

type Server struct {
	Handler Handler
	Logger  Logger

	childIndex int
	conn       *net.UDPConn
}

func (s *Server) ListenAndServe(addr string) error {
	if s.childIndex == 0 {
		return s.spwan(addr)
	}

	conn, err := ListenUDP("udp", addr)
	if err != nil {
		s.Logger.Printf("dnsserver(%d) listen on addr=%s failed: %+v", s.childIndex, addr, err)
		return err
	}
	s.conn = conn

	s.Logger.Printf("dnsserver(%d) pid(%d) serving dns on %s", s.childIndex, os.Getpid(), conn.LocalAddr())

	pool := newGoroutinePool(1 * time.Minute)
	for {
		b := AcquireByteBuffer()

		b.B = b.B[:cap(b.B)]
		n, addr, err := conn.ReadFromUDP(b.B)
		b.B = b.B[:n]

		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		pool.Go(func() {
			defer ReleaseByteBuffer(b)

			req := AcquireRequest()
			defer ReleaseRequest(req)

			err := ParseRequest(b.B, req)
			if err != nil {
				return
			}

			s.Handler.ServeDNS(&responseWriter{conn, addr}, req)
		})
	}

	return nil
}

func (s *Server) spwan(addr string) (err error) {
	type racer struct {
		index int
		err   error
	}

	maxProcs := runtime.GOMAXPROCS(0)
	if runtime.GOOS != "linux" {
		maxProcs = 1
	}

	ch := make(chan racer, maxProcs)

	for i := 1; i < runtime.NumCPU(); i++ {
		go func(index int) {
			server := &Server{
				Handler:    s.Handler,
				Logger:     s.Logger,
				childIndex: i,
			}
			err := server.ListenAndServe(addr)
			ch <- racer{index, err}
		}(i)
	}

	var exited int
	for sig := range ch {
		s.Logger.Printf("one of the child workers exited with error: %v", sig.err)

		if exited++; exited > 200 {
			s.Logger.Printf("child workers exit too many times(%d)", exited)
			err = errors.New("child workers exit too many times")
			break
		}

		go func(index int) {
			server := &Server{
				Handler:    s.Handler,
				Logger:     s.Logger,
				childIndex: index,
			}
			err := server.ListenAndServe(addr)
			ch <- racer{index, err}
		}(sig.index)
	}

	return
}