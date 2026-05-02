package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	if err := runMigrations(databaseURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer pool.Close()

	handler := newAPIHandler(pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/weather", handler.weather)
	mux.HandleFunc("/api/favorites", handler.favorites)
	mux.HandleFunc("/api/favorites/", handler.favoriteByID)
	mux.HandleFunc("/api/health", healthHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      cors(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Weather backend listening on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}

func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

type apiHandler struct {
	pool *pgxpool.Pool
}

func newAPIHandler(pool *pgxpool.Pool) *apiHandler {
	return &apiHandler{pool: pool}
}

func (h *apiHandler) weather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "city query parameter is required", http.StatusBadRequest)
		return
	}

	weather, err := fetchWeather(city)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not fetch weather: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weather)
}

func (h *apiHandler) favorites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		listFavorites(w, r, h.pool)
	case http.MethodPost:
		createFavorite(w, r, h.pool)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *apiHandler) favoriteByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idParam := r.URL.Path[len("/api/favorites/"):]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid favorite id", http.StatusBadRequest)
		return
	}

	if err := deleteFavorite(r.Context(), h.pool, id); err != nil {
		http.Error(w, fmt.Sprintf("could not delete favorite: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func listFavorites(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
	rows, err := pool.Query(r.Context(), "SELECT id, city, latitude, longitude, created_at FROM favorites ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var favorites []Favorite
	for rows.Next() {
		var f Favorite
		if err := rows.Scan(&f.ID, &f.City, &f.Latitude, &f.Longitude, &f.CreatedAt); err != nil {
			http.Error(w, fmt.Sprintf("db scan error: %v", err), http.StatusInternalServerError)
			return
		}
		favorites = append(favorites, f)
	}

	json.NewEncoder(w).Encode(favorites)
}

func createFavorite(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
	var payload struct {
		City string `json:"city"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if payload.City == "" {
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	geo, err := geocodeCity(payload.City)
	if err != nil {
		http.Error(w, fmt.Sprintf("geocoding failed: %v", err), http.StatusInternalServerError)
		return
	}

	row := pool.QueryRow(r.Context(), "INSERT INTO favorites (city, latitude, longitude) VALUES ($1, $2, $3) RETURNING id, created_at", geo.Name, geo.Latitude, geo.Longitude)
	var favorite Favorite
	favorite.City = geo.Name
	favorite.Latitude = geo.Latitude
	favorite.Longitude = geo.Longitude
	if err := row.Scan(&favorite.ID, &favorite.CreatedAt); err != nil {
		http.Error(w, fmt.Sprintf("db insert error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(favorite)
}

func deleteFavorite(ctx context.Context, pool *pgxpool.Pool, id int) error {
	cmd, err := pool.Exec(ctx, "DELETE FROM favorites WHERE id = $1", id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("favorite not found")
	}
	return nil
}

func fetchWeather(city string) (*WeatherResponse, error) {
	geo, err := geocodeCity(city)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true&timezone=auto", geo.Latitude, geo.Longitude)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			Windspeed   float64 `json:"windspeed"`
			Weathercode int     `json:"weathercode"`
			Time        string  `json:"time"`
		} `json:"current_weather"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &WeatherResponse{
		City:        geo.Name,
		Temperature: apiResp.CurrentWeather.Temperature,
		WindSpeed:   apiResp.CurrentWeather.Windspeed,
		WeatherCode: apiResp.CurrentWeather.Weathercode,
		Time:        apiResp.CurrentWeather.Time,
	}, nil
}

type Favorite struct {
	ID        int       `json:"id"`
	City      string    `json:"city"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
}

type WeatherResponse struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	WindSpeed   float64 `json:"windSpeed"`
	WeatherCode int     `json:"weatherCode"`
	Time        string  `json:"time"`
}

type geoResult struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
}

type geocodeResponse struct {
	Results []geoResult `json:"results"`
}

func geocodeCity(city string) (*geoResult, error) {
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", city)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result geocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no geocoding results for %s", city)
	}
	return &result.Results[0], nil
}
