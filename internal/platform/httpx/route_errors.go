package httpx

import "net/http"

func RouteErrors(mux *http.ServeMux) http.Handler {
	return Handle(func(w http.ResponseWriter, r *http.Request) (*Response, error) {
		h, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return nil, nil
		}

		probe := &headerProbe{header: make(http.Header)}
		h.ServeHTTP(probe, r)

		err := NotFoundError(CodeRouteNotFound, MsgRouteNotFound)
		if probe.status == http.StatusMethodNotAllowed {
			w.Header().Set("Allow", probe.header.Get("Allow"))
			err = MethodNotAllowedError()
		}

		return nil, err
	})
}

type headerProbe struct {
	header http.Header
	status int
}

func (p *headerProbe) Header() http.Header         { return p.header }
func (p *headerProbe) WriteHeader(status int)      { p.status = status }
func (p *headerProbe) Write(b []byte) (int, error) { return len(b), nil }
