package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Set gin to test mode to reduce noise in test output
	gin.SetMode(gin.TestMode)
}

func TestRouterBasicFunctionality(t *testing.T) {
	router := New()
	server := httptest.NewServer(router)
	defer server.Close()

	testClient := &http.Client{}

	t.Run("Missing bill parameter", func(t *testing.T) {
		// Make request without bill ID
		url := server.URL + "/v1//cal.ics?token=test-token"
		resp, err := testClient.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return bad request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for missing bill")
		
		// Read response body
		buf := make([]byte, 256)
		n, _ := resp.Body.Read(buf)
		content := string(buf[:n])
		assert.Contains(t, content, "missing bill", "Should return missing bill error")
	})

	t.Run("Missing token parameter", func(t *testing.T) {
		// Make request without token
		url := server.URL + "/v1/12345678/cal.ics"
		resp, err := testClient.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return bad request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for missing token")
		
		// Read response body
		buf := make([]byte, 256)
		n, _ := resp.Body.Read(buf)
		content := string(buf[:n])
		assert.Contains(t, content, "missing token", "Should return missing token error")
	})
}

func TestRouterHTTPMethods(t *testing.T) {
	router := New()
	server := httptest.NewServer(router)
	defer server.Close()

	testClient := &http.Client{}

	t.Run("POST method not allowed", func(t *testing.T) {
		url := server.URL + "/v1/12345678/cal.ics?token=test-token"
		req, _ := http.NewRequest("POST", url, nil)
		resp, err := testClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for POST")
	})

	t.Run("PUT method not allowed", func(t *testing.T) {
		url := server.URL + "/v1/12345678/cal.ics?token=test-token"
		req, _ := http.NewRequest("PUT", url, nil)
		resp, err := testClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for PUT")
	})

	t.Run("DELETE method not allowed", func(t *testing.T) {
		url := server.URL + "/v1/12345678/cal.ics?token=test-token"
		req, _ := http.NewRequest("DELETE", url, nil)
		resp, err := testClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for DELETE")
	})
}

func TestRouterStructure(t *testing.T) {
	t.Run("Router creation", func(t *testing.T) {
		router := New()
		assert.NotNil(t, router, "Should create router successfully")
	})

	t.Run("Router returns gin engine", func(t *testing.T) {
		router := New()
		assert.IsType(t, &gin.Engine{}, router, "Should return gin.Engine")
	})
}
