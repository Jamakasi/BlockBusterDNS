package cut_cname

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("cut_cname")

// dns64_hack performs dns64_hack.
type CutCname struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (s CutCname) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	nw := nonwriter.New(w)
	rcode, err := plugin.NextOrFailure(s.Name(), s.Next, ctx, nw, r)
	if err != nil {
		return rcode, err
	}

	r = nw.Msg
	if r == nil {
		return 1, fmt.Errorf("no answer received")
	}

	ret := dns.Msg{}
	ret.SetReply(r)
	ret.Truncated = r.Truncated
	ret.Extra = r.Extra
	ret.Ns = r.Ns
	ret.Answer = make([]dns.RR, 0)
	resrr := s.resolveCname(r.Question[0].Name, r.Answer)
	fmt.Println(resrr)
	if resrr != nil {
		ret.Answer = append(ret.Answer, resrr)
		w.WriteMsg(&ret)
		return ret.MsgHdr.Rcode, nil
	}

	w.WriteMsg(r)
	//fmt.Println("responce return")
	return r.MsgHdr.Rcode, nil
}

func (s CutCname) resolveCname(d string, rr []dns.RR) dns.RR {
	for _, record := range rr {
		r_header := record.Header()
		if r_header.Name == d {
			if (r_header.Rrtype == dns.TypeA) || (r_header.Rrtype == dns.TypeAAAA) {
				return record
			}
			if r_header.Rrtype == dns.TypeCNAME {
				cname := record.(*dns.CNAME).Target
				res := s.resolveCname(cname, rr)
				if res != nil {
					if res.Header().Rrtype == dns.TypeAAAA {
						resolved := &dns.AAAA{
							Hdr: dns.RR_Header{
								Name:   d,
								Rrtype: dns.TypeAAAA,
								Class:  res.Header().Class,
								Ttl:    res.Header().Ttl,
							},
							AAAA: res.(*dns.AAAA).AAAA,
						}
						return resolved
					}
					if res.Header().Rrtype == dns.TypeA {
						resolved := &dns.A{
							Hdr: dns.RR_Header{
								Name:   d,
								Rrtype: dns.TypeA,
								Class:  res.Header().Class,
								Ttl:    res.Header().Ttl,
							},
							A: res.(*dns.A).A,
						}
						return resolved
					}
				}
				return nil
			}
		}
	}
	return nil
}

// Name implements the Handler interface.
func (s CutCname) Name() string { return "cut_cname" }
