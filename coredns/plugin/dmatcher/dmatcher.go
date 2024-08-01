package dmatcher

import (
	"context"

	"github.com/coredns/coredns/plugin"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

type DConf struct {
	Port            string
	jumpToDNS       string
	storageType     string
	storegeTo       string
	storageInstance IStorage
	notifyOther     []string
	log             clog.P
}

// DomainMatcher is a basic request logging plugin.
type DMatcher struct {
	Next plugin.Handler
	Conf *DConf
}

// ServeDNS implements the plugin.Handler interface.
func (dm DMatcher) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	contain, _ := dm.Conf.storageInstance.ContainDomain(r.Question[0].Name)
	dm.Conf.log.Debugf("query storage for domain: %s. It is in db: %t", r.Question[0].Name, contain)
	if contain {
		c := new(dns.Client)
		resp, _, err := c.Exchange(r, dm.Conf.jumpToDNS)
		if err != nil {
			dm.Conf.log.Errorf("failed exchange to next dns chain: %s\n", err)
		} else {
			w.WriteMsg(resp)
		}
		return resp.MsgHdr.Rcode, err
	}
	return plugin.NextOrFailure(dm.Name(), dm.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (dm DMatcher) Name() string { return "dmatcher" }
