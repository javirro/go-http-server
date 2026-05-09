package middleware

import "net/http"

// Chain combines multiple middleware into a single one. Middleware is applied
// in the order provided (first in the slice = outermost wrapper).
//
// Example:
//
//	handler = middleware.Chain(mw1, mw2, mw3)(baseHandler)
//	// request flow: mw1 → mw2 → mw3 → baseHandler
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
