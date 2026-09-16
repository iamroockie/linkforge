package httpx

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter

	status  int
	written int64
	wrote   bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (w *ResponseWriter) WriteHeader(status int) {
	if !w.wrote {
		w.status = status
		w.wrote = true
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.wrote = true

	n, err := w.ResponseWriter.Write(b)
	w.written += int64(n)

	return n, err
}

func (w *ResponseWriter) Status() int                 { return w.status }
func (w *ResponseWriter) Written() int64              { return w.written }
func (w *ResponseWriter) Wrote() bool                 { return w.wrote }
func (w *ResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
