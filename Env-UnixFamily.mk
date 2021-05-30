DIST_FOLDER?=$(PWD)/dist
DIST_BINARY_NAME?=dbfs
DIST_BINARY_FILE=$(DIST_FOLDER)/$(DIST_NAME)

$(DIST_FOLDER):
	mkdir -p $(DIST_FOLDER)


## Support Apps Env
MINIO_DOCKER_VOLUME?=$(PWD)/localfiles/minio-data
$(MINIO_DOCKER_VOLUME):
	mkdir -p $(MINIO_DOCKER_VOLUME)

MINIO_DOCKER_CLIENT_VOLUME?=$(PWD)/localfiles/minio-client/root-home
$(MINIO_DOCKER_CLIENT_VOLUME):
	mkdir -p $(MINIO_DOCKER_CLIENT_VOLUME)
