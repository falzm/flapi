package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

type middleware struct {
	ignored *mux.Router
}

func newMiddleware(ignored []string) *middleware {
	var mw middleware

	mw.ignored = mux.NewRouter()

	for _, path := range ignored {
		mw.ignored.PathPrefix(path)
	}

	return &mw
}

func (mw *middleware) isIgnored(r *http.Request) bool {
	var routeMatch mux.RouteMatch
	return mw.ignored.Match(r, &routeMatch)
}
