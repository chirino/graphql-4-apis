package apis

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"
)

// HandlerTransport is a http.RoundTripper implemented using an http.Handler bypassing the
// the need to go through the http/tcp stack when your http client wants to call a handler.
type HandlerTransport struct {
	http.Handler
}

func (h HandlerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	reader, writer := io.Pipe()
	resp := &response{
		channel: make(chan *http.Response, 1),
		header:  http.Header{},
		reader:  reader,
		writer:  writer,
	}
	go func() {
		h.ServeHTTP(resp, request)
		resp.WriteHeader(200)
		_ = writer.Close()
	}()
	select {
	case <-request.Context().Done():
		return nil, context.Canceled
	case response := <-resp.channel:
		response.Request = request
		return response, nil
	}
}

type response struct {
	sendCounter int32
	channel     chan *http.Response
	header      http.Header
	statusCode  int
	reader      *io.PipeReader
	writer      *io.PipeWriter
}

func (t *response) Header() http.Header {
	return t.header
}

func (t *response) WriteHeader(statusCode int) {
	// The header can only be sent once...
	if atomic.CompareAndSwapInt32(&t.sendCounter, 0, 1) {
		// copy the header just in case the handler keeps modifying it async..
		header := http.Header{}
		for k, v := range t.header {
			header[k] = v
		}
		t.channel <- &http.Response{
			Status:           http.StatusText(t.statusCode),
			StatusCode:       statusCode,
			Proto:            "http",
			ProtoMajor:       1,
			ProtoMinor:       1,
			Header:           header,
			Body:             t.reader,
			ContentLength:    -1,
			TransferEncoding: nil,
			Close:            true,
			Uncompressed:     false,
			Trailer:          nil,
			TLS:              nil,
		}
		close(t.channel)
	}
}

func (t *response) Write(bytes []byte) (int, error) {
	t.WriteHeader(http.StatusOK)
	return t.writer.Write(bytes)
}
