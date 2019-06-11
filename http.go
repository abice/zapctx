package zapctx

import (
	"net/http"
)

// HTTPLevelChangeFunc allows you to host an endpoint that will change the logging level on the default logger
func HTTPLevelChangeFunc(w http.ResponseWriter, r *http.Request) {
	zapConfig.Level.ServeHTTP(w, r)
}
