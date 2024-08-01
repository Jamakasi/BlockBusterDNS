package dns64_hack

import (
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const pluginName = "dns64_hack"

func init() { plugin.Register(pluginName, setup) }

func setup(c *caddy.Controller) error {
	dns64_hack, err := dns64_hackParse(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		dns64_hack.Next = next
		return dns64_hack
	})

	return nil
}

func dns64_hackParse(c *caddy.Controller) (*dns64_hack, error) {
	_, defaultPref, _ := net.ParseCIDR("64:ff9b::/96")
	dns64_hack := &dns64_hack{
		v6_delete:     false,
		v4_delete:     false,
		dnssec_delete: false,
		Prefix:        defaultPref,
	}

	for c.Next() {
		args := c.RemainingArgs()
		if len(args) == 1 {
			pref, err := parsePrefix(c, args[0])

			if err != nil {
				return nil, err
			}
			dns64_hack.Prefix = pref
			continue
		}
		if len(args) > 0 {
			return nil, c.ArgErr()
		}

		for c.NextBlock() {
			switch c.Val() {
			case "prefix":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				pref, err := parsePrefix(c, c.Val())

				if err != nil {
					return nil, err
				}
				dns64_hack.Prefix = pref
			case "jump-to-dns":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				dns64_hack.jumpToDNS = c.Val()
			case "v6_delete":
				dns64_hack.v6_delete = true
			case "v4_delete":
				dns64_hack.v4_delete = true
			case "dnssec_delete":
				dns64_hack.dnssec_delete = true
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}
	return dns64_hack, nil
}

func parsePrefix(c *caddy.Controller, addr string) (*net.IPNet, error) {
	_, pref, err := net.ParseCIDR(addr)
	if err != nil {
		return nil, err
	}

	// Test for valid prefix
	n, total := pref.Mask.Size()
	if total != 128 {
		return nil, c.Errf("invalid netmask %d IPv6 address: %q", total, pref)
	}
	if n%8 != 0 || n < 32 || n > 96 {
		return nil, c.Errf("invalid prefix length %q", pref)
	}

	return pref, nil
}
