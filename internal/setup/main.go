package setup

import (
	"github.com/samber/do"
	"github.com/therealpaulgg/recipe-book-proxy/pkg/cache"
)

func SetupServices(i *do.Injector) {
	do.Provide(i, cache.NewRedisClient)
}