package shop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"cashmachine/internal/config"
)

const maxBody = 20 << 20

var errNotCSV = errors.New("pas un csv")

type Outcome struct {
	Site config.Site
	Rows [][]string
	Err  error
}

func FetchAll(cfg config.Config, sites []config.Site, start, end time.Time) []Outcome {
	outcomes := make([]Outcome, len(sites))
	var wg sync.WaitGroup
	for i, site := range sites {
		wg.Add(1)
		go func(i int, site config.Site) {
			defer wg.Done()
			rows, err := fetchOne(cfg, site, start, end)
			outcomes[i] = Outcome{Site: site, Rows: rows, Err: err}
		}(i, site)
	}
	wg.Wait()
	return outcomes
}

func fetchOne(cfg config.Config, site config.Site, start, end time.Time) ([][]string, error) {
	endpoint := fmt.Sprintf(
		"http://%s%s?startdate=%s&enddate=%s",
		site.IP,
		cfg.CSVPath,
		url.QueryEscape(start.Format("02/01/2006")),
		url.QueryEscape(end.Format("02/01/2006")),
	)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "cashmachine-web/1.0")

	client := &http.Client{Timeout: cfg.Timeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("réponse HTTP %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBody {
		return nil, errors.New("fichier trop volumineux")
	}
	return ParseCSV(body)
}

func PublicError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, errNotCSV) {
		if errors.Is(err, errNotCSV) {
			return "la caisse n'a pas renvoyé un fichier de transactions"
		}
		return "délai dépassé"
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return "délai dépassé"
		}
		return "connexion impossible"
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return "connexion impossible"
	}
	message := err.Error()
	if len(message) > 160 {
		return "lecture impossible"
	}
	return message
}
