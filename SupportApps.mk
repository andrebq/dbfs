.PHONY: minio-docker

minio-docker-start: $(MINIO_DOCKER_VOLUME) $(MINIO_DOCKER_CLIENT_VOLUME)
	docker run \
		--rm \
		-d \
		-e MINIO_ROOT_USER=$(MINIO_DOCKER_USERNAME) \
		-e MINIO_ROOT_PASSWORD=$(MINIO_DOCKER_PASSWORD) \
		-p $(MINIO_DOCKER_LOCAL_PORT):9000 \
		--name $(MINIO_DOCKER_CONTAINER_NAME) \
		-v $(MINIO_DOCKER_VOLUME):/data minio/minio \
		server /data/disks{1..4}
	docker run \
		-d \
		--name $(MINIO_DOCKER_CONTAINER_NAME_CLIENT) \
		--network=container:$(MINIO_DOCKER_CONTAINER_NAME) \
		-v $(MINIO_DOCKER_CLIENT_VOLUME):/root \
		-it \
		--entrypoint /bin/bash \
		--rm \
		minio/mc \
		-c 'tail -f /dev/null'
	docker logs -f $(MINIO_DOCKER_CONTAINER_NAME)

minio-docker-stop:
	docker kill $(MINIO_DOCKER_CONTAINER_NAME_CLIENT)
	docker kill $(MINIO_DOCKER_CONTAINER_NAME)

minio-docker-client:
	docker exec -ti \
		$(MINIO_DOCKER_CONTAINER_NAME_CLIENT) \
		bash
