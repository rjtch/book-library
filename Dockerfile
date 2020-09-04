books-api:
	docker build \
		-f Dockerfile-books \
		-t $(PROJECT)/books-api \
		--build-arg PACKAGE_NAME=book-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f Dockerfile-metrics \
		-t $(PROJECT)/metrics \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidebar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.