.PHONY: generate-random-blob

randomTestFile=$(LOCALFILES_RANDOM_BLOB)/random.blob
changedTestFile=$(LOCALFILES_RANDOM_BLOB)/random-changed.blob
diffChunksOriginal=$(LOCALFILES_RANDOM_BLOB)/refs.original
diffChunksChanged=$(LOCALFILES_RANDOM_BLOB)/refs.changed
$(randomTestFile): $(LOCALFILES_RANDOM_BLOB)
	dd if=/dev/random of=$(randomTestFile) bs=1024 count=50000

# change the first few bytes of the file
# copy the tail from the original random file
$(changedTestFile): $(LOCALFILES_RANDOM_BLOB) $(randomTestFile)
	dd if=/dev/random of=$(changedTestFile) bs=1024 count=50
	dd if=$(randomTestFile) of=$(changedTestFile) bs=1024 iseek=50 oseek=50 count=49950

file?=$(randomTestFile)
print-chunks: dist $(file)
	./dist/dbfs -o human blob chunks -i $(file)

diff-chunks: dist $(randomTestFile) $(changedTestFile)
	./dist/dbfs -o json-lines blob chunks -i $(randomTestFile) | jq '.chunks[].ref' > $(diffChunksOriginal)
	./dist/dbfs -o json-lines blob chunks -i $(changedTestFile) | jq '.chunks[].ref' > $(diffChunksChanged)
	diff -u $(diffChunksOriginal) $(diffChunksChanged)

change-bytes: $(changedTestFile)
