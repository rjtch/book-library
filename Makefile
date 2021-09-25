SHELL := /bin/bash
images = $(docker images)

all:  books-api metrics

keys:
	go run ./cmd/admin/main.go keygen private.pem

admin:
	go run ./cmd/admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

migrate:
	go run ./cmd/admin/main.go --db-disable-tls=1 migrate

seed: migrate
	go run ./cmd/admin/main.go --db-disable-tls=1 seed

books-api:
	docker build \
		-f Dockerfile-books \
		-t book-api-kit \
		--build-arg PACKAGE_NAME=book-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f Dockerfile-metrics \
		-t book-metrics-kit \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidebar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

start: ##start everything with docker-compose
	 docker-compose up

up: ## Start everything with docker-compose without building
	docker-compose up

down:
	docker-compose down

test:
	go test -mod=vendor ./... -count=1

clean:
	docker system prune -f
	docker volume prune -f

rmi:
	docker rmi -f $(images)