package proxy

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/kubernetes/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"github.com/kubernetes/kubernetes/pkg/proxy/config"
	"github.com/kubernetes/kubernetes/pkg/util"
	log "github.com/cihub/seelog"
)

type serviceInfo struct {
	labels     map[string]string
	portalIP   net.IP
	portalPort int
	protocol   api.Protocol
	proxyPort  int

	mu     sync.Mutex // protects active
	active bool
}

type KubernetesSync struct {
	nclient *nginxProxy
	filter  func(map[string]string) bool

	mu         sync.Mutex // protects serviceMap
	serviceMap map[string]*serviceInfo
}

var (
	namespace    = "default"
	syncInterval = time.Second * 10
)

func newClient() (*client.Client, error) {
	protocol := os.Getenv("KUBERNETES_API_PROTOCOL")
	host := os.Getenv("KUBERNETES_RO_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_RO_SERVICE_PORT")

	if len(protocol) == 0 {
		protocol = "http"
	}

	if len(host) == 0 {
		host = "127.0.0.1"
	}

	if len(port) == 0 {
		port = "8080"
	}

	return client.New(&client.Config{
		Host: protocol + "://" + host + ":" + port,
	})
}

func (ksync *KubernetesSync) OnUpdate(services []api.Service) {
	activeServices := util.StringSet{}
	for _, service := range services {
		if !ksync.filter(service.ObjectMeta.Labels) {
			continue
		}

		activeServices.Insert(service.Name)
		info, exists := ksync.getServiceInfo(service.ObjectMeta.Name)
		serviceIP := net.ParseIP(service.Spec.PortalIP)
		if exists && (info.portalPort != service.Spec.Port || !info.portalIP.Equal(serviceIP)) {
			err := ksync.removeServer(service.ObjectMeta.Name, info)
			if err != nil {
				log.Debugf("failed to remove proxy for %q: %s", service.ObjectMeta.Name, err)
			}
		}
		log.Debugf("adding new service %q at %s:%d/%s (local :%d)", service.ObjectMeta.Name, serviceIP, service.Spec.Port, service.Spec.Protocol, service.Spec.ProxyPort)
		si := &serviceInfo{
			labels:    service.ObjectMeta.Labels,
			proxyPort: service.Spec.ProxyPort,
			protocol:  service.Spec.Protocol,
			active:    true,
		}
		ksync.setServiceInfo(service.ObjectMeta.Name, si)
		si.portalIP = serviceIP
		si.portalPort = service.Spec.Port
		err := ksync.addServer(service.ObjectMeta.Name, si)
		if err != nil {
			log.Debugf("failed to add proxy %q: %s", service.ObjectMeta.Name, err)
		}
	}
	ksync.mu.Lock()
	defer ksync.mu.Unlock()
	for name, info := range ksync.serviceMap {
		if !activeServices.Has(name) {
			err := ksync.removeServer(name, info)
			if err != nil {
				log.Debugf("failed to remove proxy for %q: %s", name, err)
			}
			delete(ksync.serviceMap, name)
		}
	}
}

func (ksync *KubernetesSync) getServiceInfo(service string) (*serviceInfo, bool) {
	ksync.mu.Lock()
	defer ksync.mu.Unlock()
	info, ok := ksync.serviceMap[service]
	return info, ok
}

func (ksync *KubernetesSync) setServiceInfo(service string, info *serviceInfo) {
	ksync.mu.Lock()
	defer ksync.mu.Unlock()
	ksync.serviceMap[service] = info
}

func (ksync *KubernetesSync) removeServer(service string, info *serviceInfo) error {
	// Remove nginx server config
	return ksync.nclient.del(service)
}

func (ksync *KubernetesSync) addServer(service string, info *serviceInfo) error {
	// Create nginx server config
	var ws bool
	if info.labels["websockets"] == "true" {
		ws = true
	}

	return ksync.nclient.set(service, &nginxConfig{
		Alias:     info.labels["hostname"],
		PortalIP:  info.portalIP.String(),
		Port:      info.portalPort,
		WebSocket: ws,
	})
}

func NewKubernetesSync(nc *nginxProxy, filter func(map[string]string) bool) *KubernetesSync {
	ks := &KubernetesSync{
		serviceMap: make(map[string]*serviceInfo),
		nclient:    nc,
		filter:     filter,
	}
	return ks
}

func WatchKubernetes(nclient *nginxProxy, filter func(map[string]string) bool) {
	c, err := newClient()
	if err != nil {
		panic(err.Error())
	}

	serviceConfig := config.NewServiceConfig()
	endpointsConfig := config.NewEndpointsConfig()

	config.NewSourceAPI(
		c.Services(api.NamespaceAll),
		c.Endpoints(api.NamespaceAll),
		syncInterval,
		serviceConfig.Channel("api"),
		endpointsConfig.Channel("api"),
	)

	ks := NewKubernetesSync(nclient, filter)

	// Wire skydns to handle changes to services
	serviceConfig.RegisterHandler(ks)
}
