SHELL := /bin/bash

export PROJECT = book-library-kit

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
		-f Dockerfile.books-api \
		-t $(PROJECT)/books-api \
		--build-arg PACKAGE_NAME=book-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f Dockerfile.metrics \
		-t $(PROJECT)/metrics \
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


# rmv-all:
#     docker rmi -f $(docker images -aq)
# 	docker rm -f $(docker ps -aq)
#
# deps-reset:
# 	git checkout -- go.mod
# 	go mod tidy
# 	go mod vendor
#
# deps-upgrade:
# 	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
# 	go get -t -d -v ./...
#
# deps-cleancache:
# 	go clean -modcache