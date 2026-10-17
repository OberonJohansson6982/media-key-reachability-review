package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type keyList struct {
	Keys []keyRecord `json:"keys"`
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("set INFRAI_API_KEY")
	}
	var required map[string][]string
	if err := json.Unmarshal([]byte(os.Getenv("MEDIA_STAGE_SCOPES")), &required); err != nil {
		log.Fatal("set MEDIA_STAGE_SCOPES to a JSON object of stage-to-scope lists: ", err)
	}
	c := client{http: &http.Client{Timeout: 30 * time.Second}, key: key}
	http.HandleFunc("GET /audit", func(w http.ResponseWriter, r *http.Request) {
		data, err := c.call(r.Context(), http.MethodGet, "/v1/account/keys/list", nil)
		if err != nil {
			respondError(w, err)
			return
		}
		var listing keyList
		if err := json.Unmarshal(data, &listing); err != nil {
			respondError(w, fmt.Errorf("decode key list: %w", err))
			return
		}
		rows := audit(listing.Keys, required)
		markdown := reportMarkdown(time.Now(), rows)
		// Rendering and key inventory share this credential and base URL.
		pdf, err := c.call(r.Context(), http.MethodPost, "/v1/pdf/generate", map[string]any{"markdown": markdown, "store": false})
		if err != nil {
			respondError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			Date string          `json:"date"`
			Keys []auditRow      `json:"keys"`
			PDF  json.RawMessage `json:"pdf"`
		}{time.Now().UTC().Format("2006-01-02"), rows, pdf}); err != nil {
			log.Print(err)
		}
	})
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}

func respondError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	var api *apiError
	if errors.As(err, &api) && api.status >= 400 && api.status < 500 {
		status = api.status
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": strings.TrimSpace(err.Error())})
}
