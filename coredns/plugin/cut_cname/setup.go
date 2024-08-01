package cut_cname

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {

	plugin.Register("cut_cname", func(c *caddy.Controller) error {
		c.Next()
		if c.NextArg() {
			return plugin.Error("cut_cname", c.ArgErr())
		}

		dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
			return CutCname{Next: next}
		})

		return nil
	})
}
