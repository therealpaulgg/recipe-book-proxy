package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/therealpaulgg/recipe-book-proxy/internal/setup"
	"github.com/therealpaulgg/recipe-book-proxy/pkg/web/router"
)

func main() {
	injector := do.New()
	setup.SetupServices(injector)
	err := godotenv.Load()
	if os.Getenv("NO_DOTENV") != "1" && err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
	r := router.Router(injector)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Info().Msg(fmt.Sprintf("Server started on port %s", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}
}