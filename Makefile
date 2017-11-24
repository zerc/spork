
build: 
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

rundev: 
	docker-compose -f docker-compose.yml -f docker-compose-dev.yml up web

run: 
	docker-compose up web

brun: build run
