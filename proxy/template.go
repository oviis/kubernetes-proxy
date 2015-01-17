package proxy

var nginxTemplate = `
{{ range $host := . }}
server {
	server_name {{$host.Host}} {{$host.Alias}};
	{{if $host.Ssl }}
	listen 443;
	ssl_certificate           /etc/nginx/ssl/{{$host.SslCrt}};
	ssl_certificate_key       /etc/nginx/ssl/{{$host.SslKey}};

	ssl on;
	ssl_session_cache  builtin:1000  shared:SSL:10m;
	ssl_protocols  TLSv1 TLSv1.1 TLSv1.2;
	ssl_ciphers HIGH:!aNULL:!eNULL:!EXPORT:!CAMELLIA:!DES:!MD5:!PSK:!RC4;
	ssl_prefer_server_ciphers on;
	{{ else }}
	listen 80;
	{{ end }}

	location / {

		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For add_x_forwarded_for;
		proxy_set_header Host $http_host;
		proxy_set_header X-NginX-Proxy true;

		{{if $host.Ssl }}
		proxy_redirect http:// https://;
		{{ else }}
		proxy_redirect off;
		{{ end }}

		{{if $host.WebSocket }}
		proxy_http_version 1.1;
		proxy_set_header    Upgrade     $http_upgrade;
		proxy_set_header    Connection  $connection_upgrade;
		{{ end }}

		proxy_pass http://{{$host.PortalIP}}:{{$host.Port}};
	}
}
{{ end }}
`
