package mapshandler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"chemin-du-local.bzh/graphql/internal/config"
)

func HandleAutocomplete(w http.ResponseWriter, r *http.Request) {
	mapApiKey := config.Cfg.Maps.Key

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Input              string `json:"input"`
		SessionTokenString string `json:"sessiontoken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Get("https://maps.googleapis.com/maps/api/place/autocomplete/json?input=" + req.Input + "&types=address&language=french&components=country:fr&key=" + mapApiKey + req.SessionTokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func HandlePlaceDetails(w http.ResponseWriter, r *http.Request) {
	mapApiKey := config.Cfg.Maps.Key

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlaceID            string `json:"placeID"`
		SessionTokenString string `json:"sessiontoken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Get("https://maps.googleapis.com/maps/api/place/details/json?place_id=" + req.PlaceID + "&fields=geometry,address_components&key=" + mapApiKey + "&sessiontoken=" + req.SessionTokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
