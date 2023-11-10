package routes

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"net/http"
	"net/url"
	"os"

	"encoding/json"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
)

type NutritionixBody struct {
	Query               string  `json:"query"`
	NumServings         int     `json:"num_servings,omitempty"`
	Aggregate           string  `json:"aggregate,omitempty"`
	LineDelimited       bool    `json:"line_delimited"`
	UseRawFoods         bool    `json:"use_raw_foods"`
	IncludeSubrecipe    bool    `json:"include_subrecipe,omitempty"`
	Timezone            string  `json:"timezone,omitempty"`
	ConsumedAt          string  `json:"consumed_at,omitempty"`
	Lat                 float64 `json:"lat,omitempty"`
	Lng                 float64 `json:"lng,omitempty"`
	MealType            int     `json:"meal_type,omitempty"`
	UseBrandedFoods     bool    `json:"use_branded_foods,omitempty"`
	Locale              string  `json:"locale,omitempty"`
	Taxonomy            bool    `json:"taxonomy,omitempty"`
	IngredientStatement bool    `json:"ingredient_statement,omitempty"`
	LastModified        bool    `json:"last_modified,omitempty"`
}

func getNutrition(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body NutritionixBody
		err := json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			log.Err(err).Msg("Error parsing request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Err(err).Msg("Error marshalling request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		minBodyBuffer := &bytes.Buffer{}
		json.Compact(minBodyBuffer, bodyBytes)
		hash := sha256.Sum256(minBodyBuffer.Bytes())
		redisClient := do.MustInvoke[*redis.Client](i)
		val, err := redisClient.Get(context.TODO(), string(hash[:])).Result()
		if err != nil && err != redis.Nil {
			log.Err(err).Msg("Error getting value from redis")
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if err == redis.Nil {
			headers := http.Header{}
			headers.Add("Content-Type", "application/json")
			headers.Add("x-app-id", os.Getenv("NUTRITIONIX_APP_ID"))
			headers.Add("x-app-key", os.Getenv("NUTRITIONIX_APP_KEY"))
			url, _ := url.Parse("https://trackapi.nutritionix.com/v2/natural/nutrients")
			newReq := &http.Request{
				Method: "POST",
				Header: headers,
				URL:    url,
				Body:   io.NopCloser(bytes.NewReader(minBodyBuffer.Bytes())),
			}
			client := http.DefaultClient
			res, err := client.Do(newReq)
			if err != nil || res.StatusCode != http.StatusOK {
				log.Err(err).Msg("Error getting value from nutritionix")
				w.WriteHeader(http.StatusFailedDependency)
				if err == nil {
					apiBody, err := io.ReadAll(res.Body)
					if err != nil {
						log.Err(err).Msg("Error reading response body")
						return
					}
					w.Write(apiBody)
				}
				return
			}
			// save res body to redis, then send to client
			apiBody, err := io.ReadAll(res.Body)
			if err != nil {
				log.Err(err).Msg("Error reading response body")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = redisClient.Set(context.TODO(), string(hash[:]), string(apiBody), 0).Result()
			if err != nil {
				log.Err(err).Msg("Error setting value in redis")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(apiBody)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(val))
		}
	}
}

func getItem(i *do.Injector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract the following query params: nix_item_id, upc, rw_sin, claims, taxonomy

		url := url.URL{}
		url.Scheme = "https"
		url.Host = "trackapi.nutritionix.com"
		url.Path = "/v2/search/item"
		queryVals := url.Query()
		if nixItemId := r.URL.Query().Get("nix_item_id"); nixItemId != "" {
			queryVals.Add("nix_item_id", nixItemId)
		}
		if upc := r.URL.Query().Get("upc"); upc != "" {
			queryVals.Add("upc", upc)
		}
		if rwSin := r.URL.Query().Get("rw_sin"); rwSin != "" {
			queryVals.Add("rw_sin", rwSin)
		}
		if claims := r.URL.Query().Get("claims"); claims != "" {
			queryVals.Add("claims", claims)
		}
		if taxonomy := r.URL.Query().Get("taxonomy"); taxonomy != "" {
			queryVals.Add("taxonomy", taxonomy)
		}
		url.RawQuery = queryVals.Encode()
		hash := sha256.Sum256([]byte(url.String()))
		redisClient := do.MustInvoke[*redis.Client](i)
		val, err := redisClient.Get(context.TODO(), string(hash[:])).Result()
		if err != nil && err != redis.Nil {
			log.Err(err).Msg("Error getting value from redis")
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if err == redis.Nil {
			headers := http.Header{}
			headers.Add("Content-Type", "application/json")
			headers.Add("x-app-id", os.Getenv("NUTRITIONIX_APP_ID"))
			headers.Add("x-app-key", os.Getenv("NUTRITIONIX_APP_KEY"))
			newReq := &http.Request{
				Method: "GET",
				Header: headers,
				URL:    &url,
			}
			client := http.DefaultClient
			res, err := client.Do(newReq)
			if err != nil || res.StatusCode != http.StatusOK {
				log.Err(err).Msg("Error getting value from nutritionix")
				w.WriteHeader(http.StatusFailedDependency)
				if err == nil {
					apiBody, err := io.ReadAll(res.Body)
					if err != nil {
						log.Err(err).Msg("Error reading response body")
						return
					}
					w.Write(apiBody)
				}
				return
			}
			// save res body to redis, then send to client
			apiBody, err := io.ReadAll(res.Body)
			if err != nil {
				log.Err(err).Msg("Error reading response body")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = redisClient.Set(context.TODO(), string(hash[:]), string(apiBody), 0).Result()
			if err != nil {
				log.Err(err).Msg("Error setting value in redis")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(apiBody)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(val))
		}
	}
}

func ProxyRoutes(i *do.Injector) chi.Router {
	r := chi.NewRouter()
	r.Post("/nutrition", getNutrition(i))
	r.Get("/item", getItem(i))
	return r
}
