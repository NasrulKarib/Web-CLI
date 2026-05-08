package http

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"websocket-server/internal/usecase"
)

// Router centralizes route definitions and middleware.
type Router struct {
	mux          *http.ServeMux
	shellHandler *ShellWSHandler
}

// NewRouter creates a Router with all routes and middleware wired up.
// factory is called once per WebSocket connection to create an isolated ShellService.
// webRoot is the path to the frontend assets directory (e.g., "web").
func NewRouter(factory func() *usecase.ShellService, webRoot string) (*Router, error) {
	mux := http.NewServeMux()

	templateHandler, err := NewTemplateHandler(webRoot)
	if err != nil {
		return nil, fmt.Errorf("init template handler: %w", err)
	}

	router := &Router{
		mux:          mux,
		shellHandler: NewShellWSHandler(factory),
	}

	// WebSocket endpoint.
	mux.HandleFunc("/ws", router.shellHandler.Handle)

	// HTML pages.
	mux.HandleFunc("/", templateHandler.HandleIndex)

	// Static assets.
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(filepath.Join(webRoot, "js")))))
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(filepath.Join(webRoot, "css")))))

	return router, nil
}

// ServeHTTP applies the middleware chain and delegates to the mux.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := r.corsMiddleware(r.loggingMiddleware(r.mux))
	handler.ServeHTTP(w, req)
}

func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Upgrade, Connection")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s â %s", req.Method, req.URL.Path, req.RemoteAddr)
		next.ServeHTTP(w, req)
	})
}
