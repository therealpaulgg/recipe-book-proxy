package router

import (
	"fmt"
	"os"

	"net/http"

	"github.com/go-chi/chi"

	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/recipe-book-proxy/pkg/web/router/routes"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware"
)

func Router(i *do.Injector) chi.Router {
	baseRouter := chi.NewRouter()
	baseRouter.Use(middleware.Log)
	// Add simple auth middleware to 401 requests that do not include the API_KEY env var in Authorization as 'Bearer <key>'
	if os.Getenv("API_KEY") != "" {
		baseRouter.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", os.Getenv("API_KEY")) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				h.ServeHTTP(w, r)
			})
		})
	} else {
		log.Warn().Msg("API_KEY not set, no auth will be used")
	}

	apiV1Router := chi.NewRouter()
	apiV1Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	apiV1Router.Mount("/proxy", routes.ProxyRoutes(i))
	baseRouter.Mount("/api/v1", apiV1Router)
	return baseRouter
}
