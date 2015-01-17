package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	log "github.com/cihub/seelog"
)

type nginxConfig struct {
	Host      string
	Alias     string
	PortalIP  string
	Port      int
	Ssl       bool
	SslCert   string
	SslKey    string
	WebSocket bool
}

type nginxProxy struct {
	dir    string
	domain string
}

func (n *nginxProxy) reload() error {
	log.Debug("checking nginx config")
	out, err := exec.Command("/usr/sbin/nginx", "-t").Output()
	if err != nil {
		log.Errorf("nginx check error: %v", err)
		return err
	}
	log.Debug("check output: ", string(out))

	log.Debug("reloading nginx")
	out, err = exec.Command("/usr/sbin/service", "nginx", "reload").Output()
	if err != nil {
		log.Errorf("nginx reload error: %v", err)
		return err
	}
	log.Debug("reload output: ", string(out))
	return nil
}

func (n *nginxProxy) del(name string) error {
	file := filepath.Join(n.dir, name+".conf")
	if err := os.Remove(file); err != nil {
		return err
	}
	return n.reload()
}

func (n *nginxProxy) set(name string, nc *nginxConfig) error {
	if err := os.MkdirAll(n.dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(n.dir, name+".conf"))
	if err != nil {
		return err
	}
	defer file.Close()

	// set base host
	if len(nc.Host) == 0 {
		nc.Host = fmt.Sprintf("%s.%s", name, n.domain)
	}

	tmpl, err := template.New(name).Parse(nginxTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, []*nginxConfig{nc})
	if err != nil {
		return err
	}

	return n.reload()
}

func NewNginxProxy(dir, domain string) *nginxProxy {
	return &nginxProxy{
		dir:    dir,
		domain: domain,
	}
}
