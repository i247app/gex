package gex

import "net/http"

type gexMux struct {
	mux            *http.ServeMux
	defaultHandler http.HandlerFunc
}

func (m *gexMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the route matches any registered patterns
	_, pattern := m.mux.Handler(r)
	if pattern == "" {
		// If no match, call the default handler
		m.defaultHandler(w, r)
		return
	}

	// Otherwise, use the regular ServeMux to handle the request
	m.mux.ServeHTTP(w, r)
}

func (m *gexMux) addRoute(path string, handler http.Handler, middleware ...Middleware) {
	// fmt.Printf("registering: %v\n", path)
	for _, mw := range middleware {
		handler = mw(handler).(http.HandlerFunc)
	}
	m.mux.HandleFunc(path, http.HandlerFunc(handler.ServeHTTP))
}
