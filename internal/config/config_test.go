package config

import "testing"

func TestParseSiteLine(t *testing.T) {
	site, err := ParseSiteLine("wemmel|Wemmel|10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if site.ID != "wemmel" || site.Name != "Wemmel" || site.IP != "10.0.0.1" {
		t.Fatalf("site inattendu : %+v", site)
	}

	if _, err := ParseSiteLine("Wemmel|Wemmel|10.0.0.1"); err == nil {
		t.Fatal("un identifiant avec majuscule doit être refusé")
	}
	if _, err := ParseSiteLine("wemmel|Wemmel"); err == nil {
		t.Fatal("une ligne incomplète doit être refusée")
	}
}
