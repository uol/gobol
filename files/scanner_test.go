package files_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"stash.uol.intranet/s3-log-uploader/files"
	"stash.uol.intranet/s3-log-uploader/structs"
	"stash.uol.intranet/s3-log-uploader/test"
)

//
// Test the file scanner
// @author rnojiri
//

func checkFiles(t *testing.T, scanPath string, resultFiles []*structs.File, expectedFiles []string) {

	m := map[string]bool{}

	for _, file := range resultFiles {
		m[file.Path] = true
	}

	for _, file := range expectedFiles {
		assert.True(t, m[scanPath+"/"+file])
	}
}

func testScan(t *testing.T, regexp, subpath string, minSize int64, expectedFiles []string, ignoredFiles []string) {

	s := files.NewScanner(regexp, minSize, test.GetLogger(t))

	scanPath := test.GetScanPathRoot(subpath)
	r, err := s.Scan(scanPath)
	assert.NoError(t, err)
	assert.Len(t, r.Errors, 0)
	assert.Len(t, r.Files, len(expectedFiles))
	assert.Len(t, r.Ignored, len(ignoredFiles))

	checkFiles(t, scanPath, r.Files, expectedFiles)
	checkFiles(t, scanPath, r.Ignored, ignoredFiles)
}

func TestScanAllFiles(t *testing.T) {

	expected := []string{
		"gzip/gziped-large-text.log.gz",
		"gzip/large-text.log",
		"rootfolder/subfolder/test.log",
		"rootfolder/subfolder/test.log.gz",
		"common.go",
		"rootfolder/empty.log",
		"rootfolder/small.log",
	}

	ignored := []string{}

	testScan(t, ".*", "test", 0, expected, ignored)
}

func TestScanSingleFile(t *testing.T) {

	expected := []string{
		"common.go",
	}

	ignored := []string{
		"gzip/gziped-large-text.log.gz",
		"gzip/large-text.log",
		"rootfolder/subfolder/test.log",
		"rootfolder/subfolder/test.log.gz",
		"rootfolder/empty.log",
		"rootfolder/small.log",
	}

	testScan(t, "\\.go$", "test", 0, expected, ignored)
}

func TestScanNoFiles(t *testing.T) {

	expected := []string{}

	ignored := []string{
		"common.go",
		"gzip/gziped-large-text.log.gz",
		"gzip/large-text.log",
		"rootfolder/subfolder/test.log",
		"rootfolder/subfolder/test.log.gz",
		"rootfolder/empty.log",
		"rootfolder/small.log",
	}

	testScan(t, "\\.exe$", "test", 0, expected, ignored)
}

func TestScanWithMinSize(t *testing.T) {

	expected := []string{
		"gzip/gziped-large-text.log.gz",
		"gzip/large-text.log",
		"rootfolder/subfolder/test.log",
		"rootfolder/subfolder/test.log.gz",
		"common.go",
	}

	ignored := []string{
		"rootfolder/empty.log",
		"rootfolder/small.log",
	}

	testScan(t, ".*", "test", 3, expected, ignored)
}
