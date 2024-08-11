
buildimage:
	docker build . -t cloudflare-dynamic-dns
	docker tag cloudflare-dynamic-dns:latest jbronson29/cloudflare-dynamic-dns:latest
	docker push jbronson29/cloudflare-dynamic-dns:latest		
