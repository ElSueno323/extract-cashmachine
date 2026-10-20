package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Site struct {
	ID   string
	Name string
	IP   string
}

type Config struct {
	ListenAddr string
	Timeout    time.Duration
	CSVPath    string
	Sites      []Site
}

func Load() (Config, error) {
	path, err := findEnvFile()
	if err != nil {
		return Config{}, err
	}
	if err := applyEnvFile(path); err != nil {
		return Config{}, err
	}

	listen := strings.TrimSpace(os.Getenv("LISTEN_ADDR"))
	if listen == "" {
		return Config{}, fmt.Errorf("LISTEN_ADDR manquant dans .env")
	}

	timeoutRaw := strings.TrimSpace(os.Getenv("HTTP_TIMEOUT_SECONDS"))
	seconds, err := strconv.Atoi(timeoutRaw)
	if err != nil || seconds <= 0 {
		return Config{}, fmt.Errorf("HTTP_TIMEOUT_SECONDS invalide dans .env")
	}

	csvPath := strings.TrimSpace(os.Getenv("CSV_PATH"))
	if !validCSVPath(csvPath) {
		return Config{}, fmt.Errorf("CSV_PATH invalide dans .env")
	}

	sites, err := sitesFromEnv()
	if err != nil {
		return Config{}, err
	}

	return Config{
		ListenAddr: listen,
		Timeout:    time.Duration(seconds) * time.Second,
		CSVPath:    csvPath,
		Sites:      sites,
	}, nil
}

func (c Config) SiteByID(id string) (Site, bool) {
	for _, site := range c.Sites {
		if site.ID == id {
			return site, true
		}
	}
	return Site{}, false
}

func findEnvFile() (string, error) {
	candidates := []string{".env"}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("fichier .env introuvable à côté du programme ou dans le dossier courant")
}

func applyEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("lecture du fichier .env impossible")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf(".env ligne %d : format clé=valeur attendu", lineNo)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			return fmt.Errorf(".env ligne %d : clé vide", lineNo)
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf(".env ligne %d : clé illisible", lineNo)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("lecture du fichier .env impossible")
	}
	return nil
}

func sitesFromEnv() ([]Site, error) {
	var sites []Site
	seen := map[string]bool{}
	for i := 1; i < 100; i++ {
		raw := strings.TrimSpace(os.Getenv(fmt.Sprintf("SITE_%d", i)))
		if raw == "" {
			if i == 1 {
				return nil, fmt.Errorf("SITE_1 manquant dans .env")
			}
			break
		}
		site, err := ParseSiteLine(raw)
		if err != nil {
			return nil, fmt.Errorf("SITE_%d : %w", i, err)
		}
		if seen[site.ID] {
			return nil, fmt.Errorf("SITE_%d : identifiant déjà utilisé", i)
		}
		seen[site.ID] = true
		sites = append(sites, site)
	}
	return sites, nil
}

func ParseSiteLine(raw string) (Site, error) {
	parts := strings.SplitN(raw, "|", 3)
	if len(parts) != 3 {
		return Site{}, fmt.Errorf("format attendu identifiant|Nom|adresse")
	}
	site := Site{
		ID:   strings.TrimSpace(parts[0]),
		Name: strings.TrimSpace(parts[1]),
		IP:   strings.TrimSpace(parts[2]),
	}
	if site.Name == "" || site.IP == "" || !validID(site.ID) {
		return Site{}, fmt.Errorf("identifiant, nom ou adresse incomplet")
	}
	return site, nil
}

func validID(id string) bool {
	if id == "" || len(id) > 40 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func validCSVPath(path string) bool {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "://") || strings.Contains(path, "..") {
		return false
	}
	return true
}
