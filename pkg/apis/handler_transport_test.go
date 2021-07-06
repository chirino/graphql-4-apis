package apis

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestHandlerTransport(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("hello", "world")
		w.WriteHeader(202)
		_, _ = w.Write([]byte("/test"))
	})
	handler.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
	})

	ht := HandlerTransport{handler}
	client := http.Client{Transport: ht}

	resp, err := client.Get("http://anything/test")
	require.Nil(t, err)
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "/test", string(bytes))
	assert.Equal(t, resp.StatusCode, 202)
	assert.Equal(t, resp.Header.Get("hello"), "world")

	resp, err = client.Get("http://anything/empty")
	require.Nil(t, err)
	defer resp.Body.Close()
	bytes, err = ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, "", string(bytes))
	assert.Equal(t, resp.StatusCode, 200)

	resp, err = client.Get("http://anything/badpath")
	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, 404)

}
