#!/usr/bin/env bash

SCRIPT_DIR="$(dirname "$0")"
IMAGE_TAG=latest
REGISTRY_DOMAIN=openshift-registry.adorsys.de
PROJECT_NAME=book-library

docker login -u github-image-pusher -p "$OPENSHIFT_TOKEN" $REGISTRY_DOMAIN || exit 1

while IFS="" read -r service_and_context || [ -n "$service_and_context" ]
do
    SERVICE_NAME=$(echo "$service_and_context" | cut -d"=" -f1)
    CONTEXT=$(echo "$service_and_context" | cut -d"=" -f2)
    echo "Deploying $SERVICE_NAME from $CONTEXT"

    IMAGE_NAME=$REGISTRY_DOMAIN/$PROJECT_NAME/$SERVICE_NAME:$IMAGE_TAG
    LATEST_IMAGE_NAME=$REGISTRY_DOMAIN/$PROJECT_NAME/$SERVICE_NAME:latest
    docker tag "$IMAGE_NAME" "$LATEST_IMAGE_NAME"
    docker push "$IMAGE_NAME"
    docker push "$LATEST_IMAGE_NAME"
done < "$SCRIPT_DIR/service.list"