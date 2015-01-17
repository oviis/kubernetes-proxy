FROM nginx:1.7.9

RUN rm -f /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/example_ssl.conf
RUN mkdir -p /etc/nginx/ssl /etc/nginx/location

ADD nginx.conf /etc/nginx/nginx.conf
ADD kubernetes-proxy /
ADD entrypoint.sh /

ENTRYPOINT [ "/entrypoint.sh" ]
