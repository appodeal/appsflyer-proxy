build:
	docker build -t registry.appodeal.com/appodeal/go-daemon-appsflyer-proxy -f Dockerfile .

push: 
	docker push registry.appodeal.com/appodeal/go-daemon-appsflyer-proxy
