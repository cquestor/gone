package gone

// WebEngine http引擎，将作为http.Handler
type WebEngine struct {
	router map[string]IHandler
}

// New 构造WebEngine
func New() *WebEngine {
	return &WebEngine{
		router: make(map[string]IHandler),
	}
}
