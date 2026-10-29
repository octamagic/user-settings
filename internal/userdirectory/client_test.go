package userdirectory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchUsers(t *testing.T) {
	t.Run("returns decoded users and sends URNs as a query parameter", func(t *testing.T) {
		var gotURNs string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotURNs = r.URL.Query().Get("urns")
			assert.Equal(t, "/users", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"urn":"urn:user:1","date_of_birth":"2000/01/02","country_code":"GB"},{"urn":"urn:user:2"}]`))
		}))
		defer srv.Close()

		users, err := New(srv.URL).FetchUsers(context.Background(), []string{"urn:user:1", "urn:user:2"})

		require.NoError(t, err)
		assert.Equal(t, "urn:user:1,urn:user:2", gotURNs)
		require.Len(t, users, 2)
		assert.Equal(t, "urn:user:1", users[0].URN)
		assert.Equal(t, "2000/01/02", *users[0].DateOfBirth)
		assert.Equal(t, "GB", *users[0].CountryCode)
		assert.Nil(t, users[1].DateOfBirth)
		assert.Nil(t, users[1].CountryCode)
	})

	t.Run("returns an error on a non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer srv.Close()

		_, err := New(srv.URL).FetchUsers(context.Background(), []string{"urn:user:1"})

		assert.ErrorContains(t, err, "502")
	})

	t.Run("returns an error on a malformed body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`not json`))
		}))
		defer srv.Close()

		_, err := New(srv.URL).FetchUsers(context.Background(), []string{"urn:user:1"})

		assert.ErrorContains(t, err, "decoding")
	})
}
