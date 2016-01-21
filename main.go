package main

import (
	"os"

	"github.com/oviis/kubernetes-proxy/proxy"
)

var (
	proxyDir    = "/etc/nginx/conf.d"
	proxyDomain = "proxy.local"
	proxyLabel  = "proxy"
)

func main() {
	if pd := os.Getenv("PROXY_NGINX_DIR"); len(pd) > 0 {
		proxyDir = pd
	}

	if pd := os.Getenv("PROXY_DOMAIN"); len(pd) > 0 {
		proxyDomain = pd
	}

	if pl := os.Getenv("PROXY_LABEL"); len(pl) > 0 {
		proxyDomain = pl
	}

	nc := proxy.NewNginxProxy(proxyDir, proxyDomain)

	proxy.WatchKubernetes(nc, func(labels map[string]string) bool {
		if _, ok := labels[proxyLabel]; ok {
			return true
		}

		return false
	})

	select {}
}
