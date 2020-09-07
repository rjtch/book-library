FROM golang:1.13 as build_books-api
ENV CGO_ENABLED 0
ARG VCS_REF
ARG PACKAGE_NAME
ARG PACKAGE_PREFIX

# Create a location in the container for the source code. Using the
# default GOPATH location.
RUN mkdir -p github.com/book-library

# Copy the module files first and then download the dependencies. If this
# doesn't change, we won't need to do this again in future builds.
COPY go.* github.com/book-library/
WORKDIR github.com/book-library
RUN go mod download

# Copy the source code into the container.
COPY private.pem private.pem
COPY cmd cmd
COPY internal internal

# Build the admin tool so we can have it in the container. This should change
# often so do this first.
WORKDIR github.com/book-library/cmd/${PACKAGE_PREFIX}admin
RUN go build -mod=readonly

# Build the service binary. We are doing this last since this will be different
# every time we run through this process.
WORKDIR github.com/book-library/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}
RUN go build -mod=readonly

#-------------------------------------------------------------------------------------#

# Build the Go Binary.
FROM golang:1.13 as build_metrics
ENV CGO_ENABLED 0
ARG VCS_REF
ARG PACKAGE_NAME
ARG PACKAGE_PREFIX

# Create a location in the container for the source code. Using the
# default GOPATH location.
RUN mkdir -p github.com/book-library

# Copy the module files first and then download the dependencies. If this
# doesn't change, we won't need to do this again in future builds.
COPY go.* github.com/book-library/
WORKDIR github.com/book-library
RUN go mod download

# Copy the source code into the container.
WORKDIR github.com/book-library
COPY go.* ./
COPY cmd cmd
COPY internal internal
COPY vendor vendor

# Build the service binary. We are doing this last since this will be different
# every time we run through this process.
WORKDIR github.com/book-library/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}
RUN go build -mod=readonly


# Run the Go Binary in Alpine.
FROM alpine:3.7
ARG BUILD_DATE
ARG VCS_REF
ARG PACKAGE_NAME
ARG PACKAGE_PREFIX
COPY --from=build_metrics github.com/book-library/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}/${PACKAGE_NAME} /app-library/main
WORKDIR /app-library
CMD /app-library/main