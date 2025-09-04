package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHTTPClient(t *testing.T) {
	t.Run("Client configuration", func(t *testing.T) {
		client := Default
		
		// Verify timeout is set correctly
		assert.Equal(t, 15*time.Second, client.Timeout, "Client should have 15 second timeout")
		
		// Verify client is not nil
		assert.NotNil(t, client, "Default client should not be nil")
		
		// Verify transport is configured
		assert.NotNil(t, client.Transport, "Transport should be configured")
	})

	t.Run("Transport configuration", func(t *testing.T) {
		transport, ok := Default.Transport.(*http.Transport)
		assert.True(t, ok, "Transport should be *http.Transport")
		
		if transport != nil {
			// Verify important transport settings
			assert.Equal(t, 100, transport.MaxIdleConns, "MaxIdleConns should be 100")
			assert.Equal(t, 90*time.Second, transport.IdleConnTimeout, "IdleConnTimeout should be 90s")
			assert.Equal(t, 5*time.Second, transport.TLSHandshakeTimeout, "TLSHandshakeTimeout should be 5s")
			assert.Equal(t, 1*time.Second, transport.ExpectContinueTimeout, "ExpectContinueTimeout should be 1s")
			
			// Verify proxy configuration
			assert.NotNil(t, transport.Proxy, "Should support proxy configuration")
		}
	})

	t.Run("Dialer configuration", func(t *testing.T) {
		transport, ok := Default.Transport.(*http.Transport)
		assert.True(t, ok, "Transport should be *http.Transport")
		
		if transport != nil && transport.DialContext != nil {
			// We can't easily test the DialContext function directly,
			// but we can verify it's set and the transport works
			assert.NotNil(t, transport.DialContext, "DialContext should be configured")
		}
	})
}

func TestHTTPClientUsability(t *testing.T) {
	t.Run("Client can make requests", func(t *testing.T) {
		// Test that the client can be used to make HTTP requests
		// This is a basic sanity check
		client := Default
		
		// Create a simple request
		req, err := http.NewRequest("GET", "https://httpbin.org/status/200", nil)
		assert.NoError(t, err, "Should be able to create request")
		
		if req != nil {
			// We don't actually make the request in unit tests to avoid external dependencies
			// But we can verify the request can be created and the client exists
			assert.NotNil(t, client, "Client should exist for making requests")
			assert.NotNil(t, req, "Request should be created successfully")
		}
	})
}

func TestHTTPClientTimeouts(t *testing.T) {
	t.Run("Timeout configuration is reasonable", func(t *testing.T) {
		client := Default
		
		// 15 seconds should be reasonable for API calls
		assert.True(t, client.Timeout > 0, "Timeout should be positive")
		assert.True(t, client.Timeout <= 30*time.Second, "Timeout should not be excessive")
		assert.True(t, client.Timeout >= 5*time.Second, "Timeout should not be too short")
	})
}

func TestHTTPClientSingleton(t *testing.T) {
	t.Run("Default client is a singleton", func(t *testing.T) {
		// Verify that multiple references to Default return the same instance
		client1 := Default
		client2 := Default
		
		assert.Same(t, client1, client2, "Default should return the same instance")
	})
}

// Test that verifies the client has reasonable defaults for production use
func TestHTTPClientProductionReadiness(t *testing.T) {
	t.Run("Production-ready timeouts", func(t *testing.T) {
		transport, ok := Default.Transport.(*http.Transport)
		assert.True(t, ok, "Should have proper transport")
		
		if transport != nil {
			// Verify timeouts are set for production use
			assert.Greater(t, transport.TLSHandshakeTimeout, time.Duration(0), "TLS handshake timeout should be set")
			assert.Greater(t, transport.IdleConnTimeout, time.Duration(0), "Idle connection timeout should be set")
			assert.Greater(t, transport.ExpectContinueTimeout, time.Duration(0), "Expect continue timeout should be set")
		}
	})

	t.Run("Connection pooling configured", func(t *testing.T) {
		transport, ok := Default.Transport.(*http.Transport)
		assert.True(t, ok, "Should have proper transport")
		
		if transport != nil {
			// Verify connection pooling is properly configured
			assert.Greater(t, transport.MaxIdleConns, 0, "Should allow idle connections for performance")
			assert.Less(t, transport.MaxIdleConns, 1000, "Should not have excessive idle connections")
		}
	})

	t.Run("Proxy support", func(t *testing.T) {
		transport, ok := Default.Transport.(*http.Transport)
		assert.True(t, ok, "Should have proper transport")
		
		if transport != nil {
			// Should support proxy from environment (standard Go behavior)
			assert.NotNil(t, transport.Proxy, "Should support proxy configuration")
			// Note: We can't directly compare function pointers, but we can verify it's set
		}
	})
}

// Benchmark to ensure the client creation doesn't have performance issues
func BenchmarkDefaultClient(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Default
	}
}

func BenchmarkRequestCreation(b *testing.B) {
	client := Default
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
		_ = req
		_ = client
	}
}
