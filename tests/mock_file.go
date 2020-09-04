package tests

import (
	"io"

	"github.com/stretchr/testify/mock"
)

// MockedFile is the mocked version of os.File
type MockedFile struct {
	mock.Mock

	readData []byte
}

// Read mocked version of os.File's Read
func (f *MockedFile) Read(p []byte) (int, error) {
	f.Called(p)
	if len(f.readData) != 0 {
		copyLen := copy(p, f.readData)
		f.readData = f.readData[copyLen:]
		var err error
		if copyLen < len(p) {
			err = io.EOF
		}
		return copyLen, err
	}

	return 0, io.EOF
}

// Write mocked version of os.File's Write
func (f *MockedFile) Write(p []byte) (int, error) {
	args := f.Called(p)
	return args.Get(0).(int), args.Error(1)
}

// Close mocked version of os.File's Close
func (f *MockedFile) Close() error {
	return f.Called().Error(1)
}

// SetReadData sets data to read
func (f *MockedFile) SetReadData(p []byte) {
	f.readData = p
}
