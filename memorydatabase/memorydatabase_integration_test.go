//go:build integration_tests || memorydatabase_tests

// Package memorydatabase_integration_test contains integration tests for the memorydatabase package.
// These tests require a real Redis server to be running and test actual Redis operations.
package memorydatabase

// TestReadStringIntegration tests ReadString operations with a real Redis server.
// This test verifies that the ReadString function works correctly with actual Redis operations.
func TestReadStringIntegration(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for Redis
	os.Setenv("REDIS_HOST", "172.17.0.2")
	os.Setenv("REDIS_PORT", "6379")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		ctx := context.Background()

		// Initialize the client
		initiateErr := redisClient.Initiate(ctx)
		if initiateErr != nil {
			t.Errorf("Initiate should not fail, Error was %s", initiateErr.Error())
			return
		}

		// Create MemoryDatabase instance
		memoryDatabase := NewMemoryDatabase(&redisClient)

		// Test 1: Read non-existent key
		value, found, err := memoryDatabase.ReadString(ctx, "test-nonexistent-key")
		if err != nil {
			t.Errorf("ReadString with non-existent key should not fail, Error was %s", err.Error())
		} else {
			if found {
				t.Error("ReadString with non-existent key should return found=false")
			}
			if value != "" {
				t.Errorf("ReadString with non-existent key should return empty string, got '%s'", value)
			}
		}

		// Test 2: Write a value and then read it
		testKey := "test-read-string-key"
		testValue := "test-value-123"

		writeErr := memoryDatabase.WriteString(ctx, testKey, testValue, 60) // 60 seconds TTL
		if writeErr != nil {
			t.Errorf("WriteString should not fail, Error was %s", writeErr.Error())
		} else {
			// Read the value we just wrote
			readValue, readFound, readErr := memoryDatabase.ReadString(ctx, testKey)
			if readErr != nil {
				t.Errorf("ReadString with existing key should not fail, Error was %s", readErr.Error())
			} else {
				if !readFound {
					t.Error("ReadString with existing key should return found=true")
				}
				if readValue != testValue {
					t.Errorf("ReadString should return the correct value, expected '%s', got '%s'", testValue, readValue)
				}
			}
		}

		// Test 3: Read with empty key
		emptyValue, emptyFound, emptyErr := memoryDatabase.ReadString(ctx, "")
		if emptyErr != nil {
			t.Errorf("ReadString with empty key should not fail, Error was %s", emptyErr.Error())
		} else {
			if emptyFound {
				t.Error("ReadString with empty key should return found=false")
			}
			if emptyValue != "" {
				t.Errorf("ReadString with empty key should return empty string, got '%s'", emptyValue)
			}
		}

		// Test 4: Read a key with special characters
		specialKey := "test-key-with-special-chars:!@#$%^&*()"
		specialValue := "special-value-456"

		specialWriteErr := memoryDatabase.WriteString(ctx, specialKey, specialValue, 30)
		if specialWriteErr != nil {
			t.Errorf("WriteString with special characters should not fail, Error was %s", specialWriteErr.Error())
		} else {
			specialReadValue, specialReadFound, specialReadErr := memoryDatabase.ReadString(ctx, specialKey)
			if specialReadErr != nil {
				t.Errorf("ReadString with special characters should not fail, Error was %s", specialReadErr.Error())
			} else {
				if !specialReadFound {
					t.Error("ReadString with special characters should return found=true")
				}
				if specialReadValue != specialValue {
					t.Errorf("ReadString with special characters should return correct value, expected '%s', got '%s'", specialValue, specialReadValue)
				}
			}
		}
	}
}

// TestReadStringWithTTLIntegration tests ReadString operations with TTL expiration.
// This test verifies that ReadString correctly handles expired keys.
func TestReadStringWithTTLIntegration(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for Redis
	os.Setenv("REDIS_HOST", "172.17.0.2")
	os.Setenv("REDIS_PORT", "6379")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		ctx := context.Background()

		// Initialize the client
		initiateErr := redisClient.Initiate(ctx)
		if initiateErr != nil {
			t.Errorf("Initiate should not fail, Error was %s", initiateErr.Error())
			return
		}

		// Create MemoryDatabase instance
		memoryDatabase := NewMemoryDatabase(&redisClient)

		// Test: Write a value with very short TTL and verify it expires
		ttlKey := "test-ttl-key"
		ttlValue := "ttl-test-value"

		// Write with 1 second TTL
		writeErr := memoryDatabase.WriteString(ctx, ttlKey, ttlValue, 1)
		if writeErr != nil {
			t.Errorf("WriteString with TTL should not fail, Error was %s", writeErr.Error())
		} else {
			// Read immediately - should find the value
			readValue, readFound, readErr := memoryDatabase.ReadString(ctx, ttlKey)
			if readErr != nil {
				t.Errorf("ReadString immediately after write should not fail, Error was %s", readErr.Error())
			} else {
				if !readFound {
					t.Error("ReadString immediately after write should return found=true")
				}
				if readValue != ttlValue {
					t.Errorf("ReadString should return correct value, expected '%s', got '%s'", ttlValue, readValue)
				}
			}

			// Wait for TTL to expire (2 seconds to be safe)
			time.Sleep(2 * time.Second)

			// Read after expiration - should not find the value
			expiredValue, expiredFound, expiredErr := memoryDatabase.ReadString(ctx, ttlKey)
			if expiredErr != nil {
				t.Errorf("ReadString after expiration should not fail, Error was %s", expiredErr.Error())
			} else {
				if expiredFound {
					t.Error("ReadString after expiration should return found=false")
				}
				if expiredValue != "" {
					t.Errorf("ReadString after expiration should return empty string, got '%s'", expiredValue)
				}
			}
		}
	}
}
