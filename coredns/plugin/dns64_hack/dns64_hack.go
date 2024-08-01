// Package dns64_hack implements a plugin that performs dns64_hack.
//
// See: RFC 6147 (https://tools.ietf.org/html/rfc6147)
package dns64_hack

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// UpstreamInt wraps the Upstream API for dependency injection during testing
type UpstreamInt interface {
	Lookup(ctx context.Context, state request.Request, name string, typ uint16) (*dns.Msg, error)
}

// dns64_hack performs dns64_hack.
type dns64_hack struct {
	Next          plugin.Handler
	Prefix        *net.IPNet
	v4_delete     bool
	v6_delete     bool
	dnssec_delete bool
	jumpToDNS     string
}

// ServeDNS implements the plugin.Handler interface.
func (d *dns64_hack) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	//fmt.Println("otherwise do the actual dns64_hack request and response synthesis")
	msg, err := d.Dodns64_hack(ctx, w, r)
	if err != nil {
		// err means we weren't able to even issue the A request
		// to CoreDNS upstream
		fmt.Println("err means we weren't able to even issue the A request to CoreDNS upstream")
		return dns.RcodeServerFailure, err
	}

	RequestsTranslatedCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	w.WriteMsg(msg)
	//fmt.Println("responce return")
	return msg.MsgHdr.Rcode, nil
}

// Name implements the Handler interface.
func (d *dns64_hack) Name() string { return "dns64_hack" }

// Dodns64_hack takes an (empty) response to an AAAA question, issues the A request,
// and synthesizes the answer. Returns the response message, or error on internal failure.
func (d *dns64_hack) Dodns64_hack(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (*dns.Msg, error) {
	//req := request.Request{W: w, Req: r} // req is unused
	//resp, err := d.Upstream.Lookup(ctx, req, req.Name(), dns.TypeA)
	m := new(dns.Msg)
	//m.SetQuestion(dns.Fqdn(r.Question[0].Name), dns.TypeA)
	var fakeResp *dns.Msg
	c := new(dns.Client)
	m.RecursionDesired = true
	switch r.Question[0].Qtype {
	//запросить А и отдельно АААА
	case dns.TypeAAAA, dns.TypeA:
		{
			chanRes := make(chan *dns.Msg, 1)
			if !d.v6_delete {
				maaa := new(dns.Msg)
				maaa.SetQuestion(dns.Fqdn(r.Question[0].Name), dns.TypeAAAA)
				go func(arg *dns.Msg, out chan<- *dns.Msg) {
					cc := new(dns.Client)
					var fakeRespAAAA *dns.Msg
					fakeRespAAAA, _, _ = cc.Exchange(maaa, d.jumpToDNS)
					//fmt.Println("ok goroutine")
					out <- fakeRespAAAA
				}(maaa, chanRes)
			}
			m.SetQuestion(dns.Fqdn(r.Question[0].Name), dns.TypeA)
			var err error
			fakeResp, _, err = c.Exchange(m, d.jumpToDNS)
			//fmt.Println("ok main")
			if err != nil {
				return nil, err
			}
			if !d.v6_delete {
				fakeRespAAAA := <-chanRes
				if fakeRespAAAA != nil {
					for _, rr := range fakeRespAAAA.Answer {
						header := rr.Header()
						if header.Rrtype == dns.TypeAAAA {
							fakeResp.Answer = append(fakeResp.Answer, rr)
						}
					}
				}
			}
		}
	default:
		{
			m.SetQuestion(dns.Fqdn(r.Question[0].Name), r.Question[0].Qtype)
			var err error
			fakeResp, _, err = c.Exchange(m, d.jumpToDNS)
			if err != nil {
				return nil, err
			}
		}
	}
	/*for _, rr := range fakeResp.Answer {
		fmt.Println("result " + dns.TypeToString[rr.Header().Rrtype] + " " + rr.String())
	}*/
	out := d.Synthesize(r, fakeResp)
	return out, nil
}

// Synthesize merges the AAAA response and the records from the A response
func (d *dns64_hack) Synthesize(origReq, resp *dns.Msg) *dns.Msg {
	ret := dns.Msg{}
	ret.SetReply(origReq)

	// persist truncated state of AAAA response
	ret.Truncated = resp.Truncated

	// 5.3.2: dns64_hack MUST pass the additional section unchanged
	ret.Extra = resp.Extra
	ret.Ns = resp.Ns

	// 5.1.7: The TTL is the minimum of the A RR and the SOA RR. If SOA is
	// unknown, then the TTL is the minimum of A TTL and 600
	/*SOATtl := uint32(600) // Default NS record TTL
	for _, ns := range resp.Ns {
		if ns.Header().Rrtype == dns.TypeSOA {
			SOATtl = ns.Header().Ttl
		}
	}*/

	ret.Answer = make([]dns.RR, 0, len(resp.Answer))
	// convert A records to AAAA records

	for _, rr := range resp.Answer {
		header := rr.Header()
		//fmt.Println("type: " + header.String())
		if (header.Rrtype == dns.TypeAAAA) && d.v6_delete {
			//drop AAAA records
			continue
		}
		//вырезать dnssec
		if d.dnssec_delete {
			//fmt.Println("delete dnssec: " + header.String())
			if (header.Rrtype == dns.TypeRRSIG) || (header.Rrtype == dns.TypeDNSKEY) ||
				(header.Rrtype == dns.TypeDS) || (header.Rrtype == dns.TypeNSEC) ||
				(header.Rrtype == dns.TypeNSEC3) || (header.Rrtype == dns.TypeCDNSKEY) ||
				(header.Rrtype == dns.TypeCDS) {
				continue
			}
		}
		// 5.3.3: All other RR's MUST be returned unchanged
		if header.Rrtype != dns.TypeA {
			ret.Answer = append(ret.Answer, rr)
			continue
		}
		if (header.Rrtype == dns.TypeA) && !d.v4_delete {
			//add orig A
			ret.Answer = append(ret.Answer, rr)
		}
		aaaa, _ := to6(d.Prefix, rr.(*dns.A).A)

		// ttl is min of SOA TTL and A TTL
		/*ttl := SOATtl
		if rr.Header().Ttl < ttl {
			ttl = rr.Header().Ttl
		}*/

		// Replace A answer with a dns64_hack AAAA answer
		ret.Answer = append(ret.Answer, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   header.Name,
				Rrtype: dns.TypeAAAA,
				Class:  header.Class,
				Ttl:    rr.Header().Ttl,
			},
			AAAA: aaaa,
		})
	}
	return &ret
}

// to6 takes a prefix and IPv4 address and returns an IPv6 address according to RFC 6052.
func to6(prefix *net.IPNet, addr net.IP) (net.IP, error) {
	addr = addr.To4()
	if addr == nil {
		return nil, errors.New("not a valid IPv4 address")
	}

	n, _ := prefix.Mask.Size()
	// Assumes prefix has been validated during setup
	v6 := make([]byte, 16)
	i, j := 0, 0

	for ; i < n/8; i++ {
		v6[i] = prefix.IP[i]
	}
	for ; i < 8; i, j = i+1, j+1 {
		v6[i] = addr[j]
	}
	if i == 8 {
		i++
	}
	for ; j < 4; i, j = i+1, j+1 {
		v6[i] = addr[j]
	}

	return v6, nil
}
