//go:build integration_tests || unit_tests || memorydatabase_tests || memorydatabase_unit_tests

// Package memorydatabase_test contains unit tests for the memorydatabase package.
// Tests cover both  client functionality and MemoryDatabase wrapper operations.
package memorydatabase

import (
	"context"
	"errors"
	"testing"
)

// ClientMock implements the Client interface for testing purposes.
type ClientMock struct {
	initiated bool
	errored   bool
}

// InitiateClient simulates the initialization of the client.
func (mock *ClientMock) InitiateClient() {
	mock.initiated = true
}

// IsClientInitiated returns the current state of the client's initialization.
func (mock *ClientMock) IsClientInitiated() bool {
	return mock.initiated
}

// WriteString uses the mocked  client to simulate writing a string value.
// This method delegates to the underlying mocked client's Set operation.
func (mock *ClientMock) WriteString(ctx context.Context, key string, value string, ttl int) error {
	if mock.errored {
		return errors.New("Fatal error")
	}
	return nil
}

// ReadString uses the mocked  client to simulate reading a string value.
// This method delegates to the underlying mocked client's Get operation.
// It properly handles the case where a key doesn't exist (goredis.Nil error).
func (mock *ClientMock) ReadString(ctx context.Context, key string) (string, bool, error) {
	var found bool = true
	var value string = ""

	if mock.errored {
		return value, found, errors.New("Fatal error")
	}
	//	value, getError := mock.client.Get(ctx, key).Result()
	//
	//	if getError != nil {
	//		found = false
	//		if getError == goredis.Nil {
	//			// Key doesn't exist in
	//			return value, found, nil
	//		} else {
	//			// Other error occurred
	//			return value, found, getError
	//		}
	//	}
	return value, found, nil
}

// TestWriteStringWithoutInitiateClient tests WriteString operation when the client is not initiated.
func TestWriteStringWithoutInitiateClient(t *testing.T) {

	ctx := context.Background()

	client := ClientMock{initiated: false}
	memoryDatabase := MemoryDatabase{client: &client}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

	required_error := "MemoryDatabase client is not initiated, cannot perform WriteString operation"

	if err == nil {
		t.Errorf("memoryDatabase.WriteString with non initiated client should fail.")
	} else {
		if err.Error() != required_error {
			t.Errorf("memoryDatabase.WriteString call with mock should return error \"%s\", it has returned \"%s\"", required_error, err.Error())
		}
	}
}

// TestWriteStringInitiateClientButFailing tests WriteString operation when the client is initiated but simulates an error during the write operation.
func TestWriteStringInitiateClientButFailing(t *testing.T) {

	ctx := context.Background()

	client := ClientMock{initiated: true, errored: true}
	memoryDatabase := MemoryDatabase{client: &client}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

	required_error := "Fatal error"

	if err == nil {
		t.Errorf("memoryDatabase.WriteString with mocked error should fail.")
	} else {
		if err.Error() != required_error {
			t.Errorf("memoryDatabase.WriteString call with mock should return error \"%s\", it has returned \"%s\"", required_error, err.Error())
		}
	}
}

// TestWriteString tests a non failed write
func TestWriteString(t *testing.T) {

	ctx := context.Background()

	client := ClientMock{initiated: true, errored: false}
	memoryDatabase := MemoryDatabase{client: &client}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString without mocked error shouldn't fail.")
	}
}

// TestReadStringWithoutInitiateClient tests ReadString operation when the client is not initiated.
func TestReadStringWithoutInitiateClient(t *testing.T) {

	ctx := context.Background()
	client := ClientMock{initiated: false, errored: false}

	memoryDatabase := MemoryDatabase{client: &client}

	_, _, err := memoryDatabase.ReadString(ctx, "anykey")

	required_error := "MemoryDatabase client is not initiated, cannot perform ReadString operation"

	if err == nil {
		t.Errorf("memoryDatabase.ReadString without being initiated should fail.")
	} else {
		if err.Error() != required_error {
			t.Errorf("memoryDatabase.ReadString without being ininitiated should return \"%s\", it has returned \"%s\"", required_error, err.Error())
		}
	}
}

// TestReadStringInitiatedButFailedClient tests ReadString operation when the client is initiated but mucks a failure
func TestReadStringInitiatedButFailedClient(t *testing.T) {

	ctx := context.Background()
	client := ClientMock{initiated: true, errored: true}

	memoryDatabase := MemoryDatabase{client: &client}

	_, _, err := memoryDatabase.ReadString(ctx, "anykey")

	required_error := "Fatal error"

	if err == nil {
		t.Errorf("memoryDatabase.ReadString without being initiated should fail.")
	} else {
		if err.Error() != required_error {
			t.Errorf("memoryDatabase.ReadString without being ininitiated should return \"%s\", it has returned \"%s\"", required_error, err.Error())
		}
	}
}

// TestReadString tests ReadString operation when the client and no errores are reported
func TestReadString(t *testing.T) {

	ctx := context.Background()
	client := ClientMock{initiated: true, errored: false}

	memoryDatabase := NewMemoryDatabase(&client)

	_, _, err := memoryDatabase.ReadString(ctx, "anykey")

	if err != nil {
		t.Errorf("memoryDatabase.ReadString should not fail.")
	}
}
