package filedatabackend

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/dkorittki/loago/pkg/instructor/databackend"
)

// FileDataBackend stores results as JSON encoded data in a file
// and implements the DataBackend interface.
type FileDataBackend struct {
	encoder *json.Encoder
	file    *os.File
	writer  *bufio.Writer
}

// New creates a new FileDataBackend and opens a new
// filehandle on the file specified by filepath.
// If the file specified by filepath already exists,
// new data is appended at the end of the file.
func New(filepath string) (*FileDataBackend, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)

	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)

	return &FileDataBackend{encoder, file, writer}, nil
}

// Store stores a result as json encoded data in a file
// and implements the Store method of the DataBackend interface.
func (b *FileDataBackend) Store(result *databackend.Result) error {
	return b.encoder.Encode(result)
}

// Close flushes all data from io buffer and
// closes the filehandle. It implements the Close method from the
// DataBackend.
func (b *FileDataBackend) Close() error {
	if err := b.writer.Flush(); err != nil {
		return err
	}

	return b.file.Close()
}
