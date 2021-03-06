.PHONY: generate-random-blob

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
	./dist/dbfs -o json blob chunks -i $(randomTestFile) | jq -c '.[] | {start: .start, ref: .ref}' > $(diffChunksOriginal)
	./dist/dbfs -o json blob chunks -i $(changedTestFile) | jq -c '.[] | {start: .start, ref: .ref}' > $(diffChunksChanged)
	diff -u $(diffChunksOriginal) $(diffChunksChanged) || { true; }

change-bytes: $(changedTestFile)
