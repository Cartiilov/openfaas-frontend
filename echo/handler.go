package function

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	latitude  = "35.6895"  // Example latitude
	longitude = "139.6917" // Example longitude
)

func getSecret(w http.ResponseWriter, key string) string {
	dat, err := os.ReadFile("/var/openfaas/secrets/" + key)
	if err != nil {
		http.Error(w, key+" secret not found", http.StatusInternalServerError)
		return ""
	}
	return strings.TrimSpace(string(dat))
}

func Handle(w http.ResponseWriter, r *http.Request) {
	apiKey := getSecret(w, "api-connection")
	if apiKey == "" {
		return
	}
	url := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%s&lon=%s&exclude=minutely,hourly,daily,alerts&appid=%s", latitude, longitude, apiKey)

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	weatherData, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the weather data to stdout
	fmt.Printf("weather data: %s", string(weatherData))

	response := struct {
		Headers     map[string][]string `json:"headers"`
		Environment []string            `json:"environment"`
		Weather     json.RawMessage     `json:"weather"`
	}{
		Headers:     r.Header,
		Environment: os.Environ(),
		Weather:     weatherData,
	}

	resBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the result
	w.WriteHeader(http.StatusOK)
	w.Write(resBody)
}
