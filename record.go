package fastdns

import (
	"net"
)

// AppendSRVRecord appends the SRV records to dst and returns the resulting dst.
func AppendSRVRecord(dst []byte, req *Message, ttl uint32, srvs []net.SRV) []byte {
	// SRV Records
	for _, srv := range srvs {
		length := 8 + len(srv.Target)
		// fixed size array for avoid bounds check
		answer := [...]byte{
			// NAME
			0xc0, 0x0c,
			// TYPE
			0x00, byte(TypeSRV),
			// CLASS
			byte(req.Question.Class >> 8), byte(req.Question.Class),
			// TTL
			byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
			// RDLENGTH
			byte(length >> 8), byte(length),
			// PRIOVRITY
			byte(srv.Priority >> 8), byte(srv.Priority),
			// WEIGHT
			byte(srv.Weight >> 8), byte(srv.Weight),
			// PORT
			byte(srv.Port >> 8), byte(srv.Port),
		}
		dst = append(dst, answer[:]...)
		// RDATA
		dst = EncodeDomain(dst, srv.Target)
	}

	return dst
}

// AppendNSRecord appends the NS records to dst and returns the resulting dst.
func AppendNSRecord(dst []byte, req *Message, ttl uint32, nameservers []net.NS) []byte {
	// NS Records
	for _, ns := range nameservers {
		// fixed size array for avoid bounds check
		answer := [...]byte{
			// NAME
			0xc0, 0x0c,
			// TYPE
			0x00, byte(TypeNS),
			// CLASS
			byte(req.Question.Class >> 8), byte(req.Question.Class),
			// TTL
			byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
			// RDLENGTH
			0x00, byte(len(ns.Host) + 2),
		}
		dst = append(dst, answer[:]...)
		// RDATA
		dst = EncodeDomain(dst, ns.Host)
	}

	return dst
}

// AppendSOARecord appends the SOA records to dst and returns the resulting dst.
func AppendSOARecord(dst []byte, req *Message, ttl uint32, mname, rname net.NS, serial, refresh, retry, expire, minimum uint32) []byte {
	length := 2 + len(mname.Host) + 2 + len(rname.Host) + 4 + 4 + 4 + 4 + 4
	// fixed size array for avoid bounds check
	answer := [...]byte{
		// NAME
		0xc0, 0x0c,
		// TYPE
		0x00, byte(TypeSOA),
		// CLASS
		byte(req.Question.Class >> 8), byte(req.Question.Class),
		// TTL
		byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
		// RDLENGTH
		byte(length >> 8), byte(length),
	}
	dst = append(dst, answer[:]...)

	// MNAME
	dst = EncodeDomain(dst, mname.Host)
	// RNAME
	dst = EncodeDomain(dst, rname.Host)

	section := [...]byte{
		// SERIAL
		byte(serial >> 24), byte(serial >> 16), byte(serial >> 8), byte(serial),
		// REFRESH
		byte(refresh >> 24), byte(refresh >> 16), byte(refresh >> 8), byte(refresh),
		// RETRY
		byte(retry >> 24), byte(retry >> 16), byte(retry >> 8), byte(retry),
		// EXPIRE
		byte(expire >> 24), byte(expire >> 16), byte(expire >> 8), byte(expire),
		// MINIMUM
		byte(minimum >> 24), byte(minimum >> 16), byte(minimum >> 8), byte(minimum),
	}
	dst = append(dst, section[:]...)

	return dst
}

// AppendMXRecord appends the MX records to dst and returns the resulting dst.
func AppendMXRecord(dst []byte, req *Message, ttl uint32, mxs []net.MX) []byte {
	// MX Records
	for _, mx := range mxs {
		length := 4 + len(mx.Host)
		// fixed size array for avoid bounds check
		answer := [...]byte{
			// NAME
			0xc0, 0x0c,
			// TYPE
			0x00, byte(TypeMX),
			// CLASS
			byte(req.Question.Class >> 8), byte(req.Question.Class),
			// TTL
			byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
			// RDLENGTH
			byte(length >> 8), byte(length),
			// PRIOVRITY
			byte(mx.Pref >> 8), byte(mx.Pref),
		}
		dst = append(dst, answer[:]...)
		// RDATA
		dst = EncodeDomain(dst, mx.Host)
	}

	return dst
}

// AppendPTRRecord appends the PTR records to dst and returns the resulting dst.
func AppendPTRRecord(dst []byte, req *Message, ttl uint32, ptr string) []byte {
	// fixed size array for avoid bounds check
	answer := [...]byte{
		// NAME
		0xc0, 0x0c,
		// TYPE
		0x00, byte(TypePTR),
		// CLASS
		byte(req.Question.Class >> 8), byte(req.Question.Class),
		// TTL
		byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
		// RDLENGTH
		00, byte(2 + len(ptr)),
	}
	dst = append(dst, answer[:]...)
	// PTR
	dst = EncodeDomain(dst, ptr)

	return dst
}

// AppendTXTRecord appends the TXT records to dst and returns the resulting dst.
func AppendTXTRecord(dst []byte, req *Message, ttl uint32, txt string) []byte {
	length := len(txt) + (len(txt)+0xff)/0x100
	// fixed size array for avoid bounds check
	answer := [...]byte{
		// NAME
		0xc0, 0x0c,
		// TYPE
		0x00, byte(TypeTXT),
		// CLASS
		byte(req.Question.Class >> 8), byte(req.Question.Class),
		// TTL
		byte(ttl >> 24), byte(ttl >> 16), byte(ttl >> 8), byte(ttl),
		// RDLENGTH
		byte(length >> 8), byte(length),
	}
	dst = append(dst, answer[:]...)

	for len(txt) > 0xff {
		// TXT Length
		dst = append(dst, 0xff)
		// TXT
		dst = append(dst, txt[:0xff]...)
		txt = txt[0xff:]
	}

	// TXT Length
	dst = append(dst, byte(len(txt)))
	// TXT
	dst = append(dst, txt...)

	return dst
}
