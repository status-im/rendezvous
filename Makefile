image:
	docker build . -t statusteam/rendezvous:latest

push:
	docker push statusteam/rendezvous:latest
