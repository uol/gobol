package files_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"stash.uol.intranet/s3-log-uploader/files"
	"stash.uol.intranet/s3-log-uploader/test"
)

//
// Test all functions from file operations package
// @author rnojiri
//

func TestGzipFile(t *testing.T) {

	filePath := test.GetScanPathRoot("test/gzip/large-text.log")

	fileInfo, err := os.Stat(filePath)
	err = files.GzipFile(filePath, filePath+".gz")
	assert.NoError(t, err)

	fileInfoGziped, err := os.Stat(filePath + ".gz")
	assert.NoError(t, err)
	assert.True(t, fileInfo.Size() > fileInfoGziped.Size())

	files.DeleteFile(filePath + ".gz")
}

func TestGzipDecompressFile(t *testing.T) {

	filePath := test.GetScanPathRoot("test/gzip/gziped-large-text.log")

	fileInfoGziped, err := os.Stat(filePath + ".gz")
	err = files.GzipDecompressFile(filePath+".gz", filePath)
	assert.NoError(t, err)

	fileInfo, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.True(t, fileInfo.Size() > fileInfoGziped.Size())

	files.DeleteFile(filePath)
}
