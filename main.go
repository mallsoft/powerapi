package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	var zones []Zone
	var lastRequest time.Time

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now().Format("15:04"), "Request received:", r.Header.Get("X-Forwarded-For"))

		if time.Now().After(lastRequest.Add(time.Minute * 5)) {
			zones = getData()
			lastRequest = time.Now()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(zones)
	})

	fmt.Println("Starting...", os.Getenv("PORT"))

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
