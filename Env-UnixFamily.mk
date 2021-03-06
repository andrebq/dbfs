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

## LocalDevTasks Env
LOCALFILES_RANDOM_BLOB?=$(PWD)/localfiles/random-blobs
$(LOCALFILES_RANDOM_BLOB):
	mkdir -p $(LOCALFILES_RANDOM_BLOB)

randomTestFile=$(LOCALFILES_RANDOM_BLOB)/random.blob
changedTestFile=$(LOCALFILES_RANDOM_BLOB)/random-changed.blob
diffChunksOriginal=$(LOCALFILES_RANDOM_BLOB)/refs.original
diffChunksChanged=$(LOCALFILES_RANDOM_BLOB)/refs.changed
