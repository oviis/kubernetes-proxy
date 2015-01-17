# Kubernetes Proxy

The kubernetes proxy is used as a fixed point of entry for external access to services run in kubernetes. 
It watches the kubernetes api for service changes and creates nginx server configs for those that have 
the label identifying them to be proxied, e.g proxy=true.

## Getting started

### Env variables

The following variables need to be set at runtime:

- PROXY_NGINX_DIR - The location where the proxy should write nginx config e.g /etc/nginx/conf.d
- PROXY_DOMAIN - The domain to proxy for e.g. set as example.com, service records will be [service].example.com
- PROXY_LABEL - The label to identify services which should be proxied e.g "proxy" will look for proxy=true

 These values will already be set when running the proxy inside kubernetes:
- KUBERNETES_RO_SERVICE_HOST - The kubernetes master read only host ip
- KUBERNETES_RO_SERVICE_PORT - The kubernetes master read only host port
- KUBERNETES_API_PROTOCOL - The kubernetes master protocol e.g http

### Running the proxy

Set the PROXY_DOMAIN in proxy.json 
Set the publicIPs field in proxy-service.json to the ips of the minions on which the proxy will run

Start the proxy
```
kubectl create -f proxy.json
kubectl create -f proxy-service.json
```

## Acknowledgements

Service watch code from SkyDNS.
