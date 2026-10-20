package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"cashmachine/internal/config"
	"cashmachine/internal/files"
	"cashmachine/internal/sheet"
	"cashmachine/internal/shop"
)

const previewLimit = 250

const xlsxType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

type API struct {
	cfg   config.Config
	store *files.Store
}

func New(cfg config.Config, store *files.Store) API {
	return API{cfg: cfg, store: store}
}

func (a API) Sites(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	sites := make([]map[string]string, 0, len(a.cfg.Sites))
	for _, site := range a.cfg.Sites {
		sites = append(sites, map[string]string{
			"id":   site.ID,
			"name": site.Name,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"defaultStart": start.Format("2006-01-02"),
		"defaultEnd":   now.Format("2006-01-02"),
		"sites":        sites,
	})
}

type exportRequest struct {
	Sites []string `json:"sites"`
	Start string   `json:"start"`
	End   string   `json:"end"`
}

type ticket struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	OK           bool       `json:"ok"`
	Transactions int        `json:"transactions"`
	Filename     string     `json:"filename"`
	Download     string     `json:"download"`
	Headers      []string   `json:"headers"`
	Rows         [][]string `json:"rows"`
	Truncated    bool       `json:"truncated"`
	Error        string     `json:"error"`
}

func (a API) Export(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "requête illisible")
		return
	}
	var payload exportRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "requête illisible")
		return
	}

	start, end, err := parseRange(payload.Start, payload.End)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(payload.Sites) == 0 {
		writeError(w, http.StatusBadRequest, "choisissez au moins une caisse")
		return
	}

	selected := make([]config.Site, 0, len(payload.Sites))
	seen := map[string]bool{}
	for _, id := range payload.Sites {
		if seen[id] {
			continue
		}
		seen[id] = true
		site, ok := a.cfg.SiteByID(id)
		if !ok {
			writeError(w, http.StatusBadRequest, "caisse inconnue")
			return
		}
		selected = append(selected, site)
	}

	outcomes := shop.FetchAll(a.cfg, selected, start, end)
	tickets := make([]ticket, 0, len(outcomes))
	var zipEntries []sheet.ZipEntry

	for _, outcome := range outcomes {
		item := ticket{
			ID:       outcome.Site.ID,
			Name:     outcome.Site.Name,
			Headers:  []string{},
			Rows:     [][]string{},
			Filename: fileName(outcome.Site.ID, start, end),
		}
		if outcome.Err != nil {
			item.Error = shop.PublicError(outcome.Err)
			tickets = append(tickets, item)
			continue
		}

		data, err := sheet.Transactions(outcome.Rows)
		if err != nil {
			item.Error = "création du fichier Excel impossible"
			tickets = append(tickets, item)
			continue
		}

		token := a.store.Put(item.Filename, xlsxType, data)
		headers, rows, truncated := sheet.Preview(outcome.Rows, previewLimit)
		item.OK = true
		item.Transactions = sheet.TransactionCount(outcome.Rows)
		item.Download = "/api/download/" + token
		item.Headers = headers
		item.Rows = rows
		item.Truncated = truncated
		tickets = append(tickets, item)
		zipEntries = append(zipEntries, sheet.ZipEntry{Name: item.Filename, Data: data})
	}

	zipDownload := ""
	if len(zipEntries) > 1 {
		zipName := fmt.Sprintf("transactions_%s_%s.zip", start.Format("02-01-2006"), end.Format("02-01-2006"))
		zipped, err := sheet.Zip(zipEntries)
		if err == nil {
			token := a.store.Put(zipName, "application/zip", zipped)
			zipDownload = "/api/download/" + token
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results":     tickets,
		"zipDownload": zipDownload,
	})
}

func (a API) Download(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	item, ok := a.store.Get(token)
	if !ok {
		writeError(w, http.StatusNotFound, "fichier introuvable ou expiré")
		return
	}
	w.Header().Set("Content-Type", item.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", item.Name))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(item.Data)
}

func parseRange(startRaw, endRaw string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("date de début invalide")
	}
	end, err := time.Parse("2006-01-02", endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("date de fin invalide")
	}
	if start.After(end) {
		return time.Time{}, time.Time{}, errors.New("la date de début doit précéder la date de fin")
	}
	return start, end, nil
}

func fileName(id string, start, end time.Time) string {
	return fmt.Sprintf("%s_%s_%s.xlsx", id, start.Format("02-01-2006"), end.Format("02-01-2006"))
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
