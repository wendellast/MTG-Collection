package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Card struct {
	Name      string   `json:"name"`
	ImageURL  string   `json:"imageUrl"`
	ManaCost  string   `json:"manaCost"`
	Colors    []string `json:"colors"`
	Type      string   `json:"type"`
	Subtypes  []string `json:"subtypes"`
	Rarity    string   `json:"rarity"`
	SetName   string   `json:"setName"`
	Languages []string `json:"languages"`
}

type Response struct {
	Cards []Card `json:"cards"`
}

func fetchCards() ([]Card, error) {

	url := "https://api.magicthegathering.io/v1/cards"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Cards, nil
}

func cardHandler(w http.ResponseWriter, r *http.Request) {
	cards, err := fetchCards()
	if err != nil {
		http.Error(w, "Erro ao buscar dados da API", http.StatusInternalServerError)
		return
	}

	nameFilter := r.URL.Query().Get("name")
	colorFilter := r.URL.Query().Get("color")
	rarityFilter := r.URL.Query().Get("rarity")
	typeFilter := r.URL.Query().Get("type")
	subtypeFilter := r.URL.Query().Get("subtype")
	setFilter := r.URL.Query().Get("set")
	languageFilter := r.URL.Query().Get("language")

	filteredCards := cards[:0]
	for _, card := range cards {
		if (nameFilter == "" || strings.Contains(strings.ToLower(card.Name), strings.ToLower(nameFilter))) &&
			(colorFilter == "" || contains(card.Colors, colorFilter)) &&
			(rarityFilter == "" || card.Rarity == rarityFilter) &&
			(typeFilter == "" || strings.Contains(strings.ToLower(card.Type), strings.ToLower(typeFilter))) &&
			(subtypeFilter == "" || contains(card.Subtypes, subtypeFilter)) &&
			(setFilter == "" || card.SetName == setFilter) &&
			(languageFilter == "" || contains(card.Languages, languageFilter)) {
			filteredCards = append(filteredCards, card)
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 40
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(filteredCards) {
		end = len(filteredCards)
	}

	tmpl := template.Must(template.ParseFiles("templates/card_list.html"))
	if err := tmpl.Execute(w, filteredCards[start:end]); err != nil {
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
		return
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	http.HandleFunc("/", cardHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
