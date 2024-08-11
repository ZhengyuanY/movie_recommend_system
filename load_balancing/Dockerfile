FROM nginx:1.27.0

COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY init.sh /etc/nginx/init.sh

CMD ["bash", "/etc/nginx/init.sh"]
