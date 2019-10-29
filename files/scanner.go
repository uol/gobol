package files

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//
// File directory scanner implementation.
// @author rnojiri
//

// Scanner - Scans a directory and find files using regular expression
type Scanner struct {
	fileRegexp   *regexp.Regexp
	minFileSize  int64
	filesFound   []*File
	ignoredFiles []*File
	errorsFound  []error
	logger       *zerolog.Logger
}

// NewScanner - builds a new Scanner
func NewScanner(fileRegexp string, minFileSize int64) *Scanner {

	logger := log.With().Str("package", "files").Logger()

	if e := logger.Info(); e.Enabled() {
		e.Str("func", "NewScanner").Msg(fmt.Sprintf("creating a new scanner using regexp '%s' and minimum file size '%d'", fileRegexp, minFileSize))
	}

	return &Scanner{
		fileRegexp:  regexp.MustCompile(fileRegexp),
		minFileSize: minFileSize,
		logger:      &logger,
	}
}

// visit - visits a file or a directory
func (s *Scanner) visit(path string, f os.FileInfo, err error) error {

	if err != nil {
		s.errorsFound = append(s.errorsFound, err)
		return err
	}

	if f.IsDir() {
		if e := s.logger.Debug(); e.Enabled() {
			e.Str("func", "visit").Msg("ignoring directory:" + path)
		}
		return nil
	}

	file := File{
		Name:             f.Name(),
		Path:             path,
		Size:             f.Size(),
		LastModification: f.ModTime(),
	}

	if s.fileRegexp.MatchString(path) {

		if e := s.logger.Debug(); e.Enabled() {
			e.Str("func", "visit").Msg("file name matches with regexp: " + path)
		}

		if file.Size < s.minFileSize {

			if e := s.logger.Debug(); e.Enabled() {
				e.Str("func", "visit").Msg(fmt.Sprintf("file does not have the minimum size: %s (%d/%d)", path, file.Size, s.minFileSize))
			}

			file.Ignored = true
			s.ignoredFiles = append(s.ignoredFiles, &file)
			return nil
		}

		file.Ignored = false
		s.filesFound = append(s.filesFound, &file)
		return nil
	}

	file.Ignored = true
	s.ignoredFiles = append(s.ignoredFiles, &file)
	return nil
}

// Scan - searches recursively for files with the same regexp pattern selected
func (s *Scanner) Scan(directoryRoot string) (*ScanResult, error) {

	s.filesFound = []*File{}
	s.ignoredFiles = []*File{}
	s.errorsFound = []error{}

	err := filepath.Walk(directoryRoot, s.visit)

	return &ScanResult{
		Files:   s.filesFound,
		Ignored: s.ignoredFiles,
		Errors:  s.errorsFound,
	}, err
}
