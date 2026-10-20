package shop

import "testing"

func TestParseCSVSemicolon(t *testing.T) {
	raw := []byte("AID;Start;totalprice\r\n1;11/09/2026;12,50\r\n\"a;b\";x;y\r\n")
	rows, err := ParseCSV(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("3 lignes attendues, obtenu %d", len(rows))
	}
	if rows[0][0] != "AID" || rows[1][2] != "12,50" {
		t.Fatalf("découpage inattendu : %#v", rows)
	}
	if rows[2][0] != "a;b" || rows[2][1] != "x" {
		t.Fatalf("guillemets mal lus : %#v", rows[2])
	}
}

func TestParseCSVRejectsHTML(t *testing.T) {
	if _, err := ParseCSV([]byte("<html>erreur</html>")); err == nil {
		t.Fatal("une page HTML doit être refusée")
	}
}
