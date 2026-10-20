package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cashmachine/internal/config"
	"cashmachine/internal/files"
)

func TestSitesDoesNotExposeAddresses(t *testing.T) {
	api := New(config.Config{
		Sites: []config.Site{{ID: "boutique", Name: "Boutique", IP: "10.1.2.3"}},
	}, files.New())

	request := httptest.NewRequest(http.MethodGet, "/api/sites", nil)
	recorder := httptest.NewRecorder()
	api.Sites(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("statut %d", recorder.Code)
	}
	body := recorder.Body.String()
	if strings.Contains(body, "10.1.2.3") {
		t.Fatal("l'adresse de la caisse ne doit pas apparaître dans la réponse")
	}
	if !strings.Contains(body, "Boutique") {
		t.Fatal("le nom de la caisse doit apparaître")
	}
}
