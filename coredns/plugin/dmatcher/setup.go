package dmatcher

import (
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

func init() { plugin.Register("dmatcher", setup) }

func setup(c *caddy.Controller) error {
	conf, err := dmatcherParse(c)

	if err != nil {
		return plugin.Error("dmatcher", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return DMatcher{Next: next, Conf: conf}
	})

	switch conf.storageType {
	case "ram-file":
		{
			conf.storageInstance = NewTree(conf.storegeTo)
		}
	case "memdb":
		{
			conf.storageInstance = NewMemDB(conf.storegeTo, conf.log)
		}
	default:
		{
			conf.storageInstance = NewTree(conf.storegeTo)
		}
	}

	//http server run
	wui := &WUI{
		Conf: conf,
	}
	wui.startWUI()
	//load from disk
	return nil
}

func dmatcherParse(c *caddy.Controller) (*DConf, error) {
	conf := &DConf{}
	conf.log = clog.NewWithPlugin("dmatcher")
	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "port":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.Port = c.Val()
			case "storage-type":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.storageType = c.Val()
			case "storage-to":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.storegeTo = c.Val()
			case "jump-to-dns":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				conf.jumpToDNS = c.Val()
			case "notify":
				{
					if !c.NextArg() {
						continue
						//return nil, c.ArgErr()
					}
					var adrr []string
					conf.notifyOther = adrr
					arr := strings.Split(c.Val(), ";")
					conf.notifyOther = append(conf.notifyOther, arr...)
				}
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}
	return conf, nil
}
