package querybuilder

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
)

type sortModel struct {
	Uuid             string `db:"uuid"`
	RegistrationNo   *int   `db:"registration_no" sorting:"registration_no,order=ASC,priority=2"`
	Lastname         string `db:"lastname" sorting:"lastname"`
	RegistrationDate string `db:"registration_date" sorting:"registration_date,order=DESC,priority=1"`
	Firstname        string `db:"firstname"`
	Hidden           string `db:"-" sorting:"hidden"`
}

func orderBy(t *testing.T, q squirrel.SelectBuilder) string {
	t.Helper()
	sql, _, err := q.ToSql()
	if err != nil {
		t.Fatal(err)
	}
	i := strings.Index(sql, "ORDER BY ")
	if i < 0 {
		return ""
	}
	return sql[i+len("ORDER BY "):]
}

func TestApplySortMapIsStable(t *testing.T) {
	sm := SortMap{"registration_no": "ASC", "registration_date": "desc", "lastname": "ASC"}
	want := "lastname ASC, registration_date DESC, registration_no ASC"
	for i := 0; i < 50; i++ {
		if got := orderBy(t, ApplySortMap(squirrel.Select("*").From("t"), sm)); got != want {
			t.Fatalf("essai %d : %q, attendu %q", i, got, want)
		}
	}
	if got := orderBy(t, ApplySortMap(squirrel.Select("*").From("t"), nil)); got != "" {
		t.Fatalf("map vide : %q", got)
	}
}

func TestApplySortMapForPriority(t *testing.T) {
	sm := SortMap{"registration_no": "ASC", "registration_date": "DESC", "lastname": "DESC"}
	// priority=1 puis priority=2, puis lastname (sans priority), puis départage.
	want := "registration_date DESC, registration_no ASC, lastname DESC, uuid ASC"
	for i := 0; i < 50; i++ {
		if got := orderBy(t, ApplySortMapFor(squirrel.Select("*").From("t"), sm, sortModel{}, "uuid")); got != want {
			t.Fatalf("essai %d : %q, attendu %q", i, got, want)
		}
	}
}

func TestApplySortMapForSafety(t *testing.T) {
	sm := SortMap{
		"registration_date":     "DESC",
		"firstname":             "desc",        // colonne du modèle sans balise sorting : acceptée, après
		"evil; DROP TABLE t --": "ASC",         // colonne inconnue : ignorée
		"registration_no":       "ASC; DROP t", // direction invalide : ASC
		"uuid":                  "DESC",        // déjà triée : le départage ne la répète pas
	}
	got := orderBy(t, ApplySortMapFor(squirrel.Select("*").From("t"), sm, &sortModel{}, "uuid"))
	want := "registration_date DESC, registration_no ASC, firstname DESC, uuid DESC"
	if got != want {
		t.Fatalf("%q, attendu %q", got, want)
	}
	// Sans tri demandé : seul le départage s'applique.
	if got := orderBy(t, ApplySortMapFor(squirrel.Select("*").From("t"), nil, sortModel{}, "uuid")); got != "uuid ASC" {
		t.Fatalf("départage seul : %q", got)
	}
}

func TestParseSortMapWithPriority(t *testing.T) {
	r := httptest.NewRequest("GET", "/?sorting_registration_no_order=desc&sorting_lastname_order=nimporte", nil)
	sm, err := ParseSortMap(sortModel{}, r)
	if err != nil {
		t.Fatal(err)
	}
	want := SortMap{"registration_no": "DESC", "registration_date": "DESC", "lastname": "ASC"}
	if len(sm) != len(want) {
		t.Fatalf("%v, attendu %v", sm, want)
	}
	for k, v := range want {
		if sm[k] != v {
			t.Fatalf("%s = %q, attendu %q (%v)", k, sm[k], v, sm)
		}
	}
	if _, err := ParseSortMap("pas une struct", r); err == nil {
		t.Fatal("erreur attendue pour un modèle qui n'est pas une struct")
	}
}
