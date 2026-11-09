package querybuilder

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/Masterminds/squirrel"
)

type Sort struct {
	By  string
	Dir string
}

func ApplySort(q squirrel.SelectBuilder, s Sort) squirrel.SelectBuilder {
	if s.By == "" {
		return q
	}
	dir := "ASC"
	if s.Dir == "desc" || s.Dir == "DESC" {
		dir = "DESC"
	}
	return q.OrderBy(fmt.Sprintf("%s %s", s.By, dir))
}

// BuildOrderByFromStruct génère une clause ORDER BY dynamique à partir d’une struct de filtre
func BuildOrderByFromStruct(q squirrel.SelectBuilder, filter any) (squirrel.SelectBuilder, error) {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return q, nil
	}

	t := v.Type()
	var orders []Sort

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Ne traiter que les champs se terminant par "Sorting"
		if !strings.HasSuffix(field.Name, "Sorting") {
			continue
		}

		dir := strings.ToLower(value.String())
		if dir != "asc" && dir != "desc" {
			continue // champ vide ou valeur invalide
		}

		// Exemple : "LastnameSorting" -> "lastname"
		column := field.Name[:len(field.Name)-len("Sorting")]
		column = camelToSnake(column)

		orders = append(orders, Sort{
			By:  column,
			Dir: dir,
		})

	}

	if len(orders) == 0 {
		return q, nil
	}

	for _, o := range orders {
		q = ApplySort(q, o)
	}

	return q, nil
}

// SortMap représente la structure d'un tri dans la map retournée
type SortMap map[string]string // clé: nom de colonne DB, valeur: direction (ASC/DESC)

// sortingField : champ d'un modèle portant une balise sorting.
//
// Format de la balise : `sorting:"<clé>,order=ASC|DESC,priority=N"`
//   - clé      : nom utilisé dans le paramètre de requête « sorting_<clé>_order » (surcharge de la direction) ;
//   - order    : direction par défaut (ASC si absente ou invalide) ;
//   - priority : rang dans la clause ORDER BY quand plusieurs colonnes sont triées (petit nombre d'abord).
//     Sans priority, c'est l'ordre de déclaration des champs qui départage.
type sortingField struct {
	key      string // clé de la balise (paramètre de requête)
	column   string // colonne SQL (balise db, sinon nom du champ en snake_case)
	order    string // ASC ou DESC
	priority int    // valeur de priority=N ; maxInt si absente
	index    int    // position du champ dans la struct
}

const noPriority = int(^uint(0) >> 1)

// modelColumn : colonne SQL d'un champ (balise db, sinon nom du champ en snake_case) ; "" si db:"-".
func modelColumn(field reflect.StructField) string {
	column := field.Tag.Get("db")
	if column == "-" {
		return ""
	}
	if column == "" {
		column = camelToSnake(field.Name)
	}
	return column
}

// structType : type struct d'un modèle (valeur ou pointeur).
func structType(model any) (reflect.Type, error) {
	t := reflect.TypeOf(model)
	if t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("modelStruct must be a struct or pointer to struct")
	}
	return t, nil
}

// sortingFields : champs du modèle portant une balise sorting, dans l'ordre de déclaration.
func sortingFields(t reflect.Type) []sortingField {
	var out []sortingField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("sorting")
		if tag == "" || tag == "-" {
			continue
		}
		parts := strings.Split(tag, ",")
		key := strings.TrimSpace(parts[0])
		column := modelColumn(field)
		if key == "" || column == "" {
			continue
		}
		f := sortingField{key: key, column: column, order: "ASC", priority: noPriority, index: i}
		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
			switch {
			case strings.HasPrefix(part, "order="):
				f.order = normalizeDirection(strings.TrimPrefix(part, "order="), "ASC")
			case strings.HasPrefix(part, "priority="):
				if n, err := strconv.Atoi(strings.TrimPrefix(part, "priority=")); err == nil {
					f.priority = n
				}
			}
		}
		out = append(out, f)
	}
	return out
}

// normalizeDirection : ASC ou DESC (insensible à la casse), sinon def.
func normalizeDirection(direction, def string) string {
	switch d := strings.ToUpper(strings.TrimSpace(direction)); d {
	case "ASC", "DESC":
		return d
	}
	return def
}

// ParseSortMap parse les query parameters en se basant sur les tags sorting d'une struct modèle
// et retourne une map structurée où :
// - La clé principale est le nom de la colonne DB (ex: "user_login")
// - La valeur est la direction du tri (ASC ou DESC)
// La direction peut être surchargée via un query parameter "sorting_{key}_order".
// Une map n'a pas d'ordre : la priorité entre colonnes (option priority=N de la balise) est appliquée par
// ApplySortMapFor, côté requête SQL.
func ParseSortMap(modelStruct interface{}, r *http.Request) (SortMap, error) {
	t, err := structType(modelStruct)
	if err != nil {
		return nil, err
	}
	qs := r.URL.Query()
	sortMap := make(SortMap)
	for _, f := range sortingFields(t) {
		order := f.order
		if override := qs.Get("sorting_" + f.key + "_order"); override != "" {
			order = normalizeDirection(override, "ASC")
		}
		sortMap[f.column] = order
	}
	return sortMap, nil
}

// ApplySortMap applique les tris d'une SortMap à une requête squirrel.SelectBuilder.
// Les colonnes sont prises par ordre alphabétique (une map Go n'a pas d'ordre : sans cela, la priorité entre
// colonnes changeait d'une requête à l'autre). Pour choisir la priorité, utiliser ApplySortMapFor.
func ApplySortMap(q squirrel.SelectBuilder, sortMap SortMap) squirrel.SelectBuilder {
	if len(sortMap) == 0 {
		return q
	}
	columns := make([]string, 0, len(sortMap))
	for column := range sortMap {
		columns = append(columns, column)
	}
	sort.Strings(columns)
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		parts = append(parts, column+" "+normalizeDirection(sortMap[column], "ASC"))
	}
	return q.OrderBy(strings.Join(parts, ", "))
}

// ApplySortMapFor applique une SortMap dans l'ordre défini par le modèle, puis ajoute les colonnes de
// départage (tiebreak, en ASC) pour un ordre total et une pagination stable.
//   - Seules les colonnes du modèle (balise db) sont acceptées : une clé inconnue est ignorée.
//   - Ordre : priority=N de la balise sorting (petit d'abord), puis ordre de déclaration des champs ;
//     les colonnes du modèle sans balise sorting viennent ensuite, par ordre alphabétique.
//   - La direction vient de la SortMap (ASC/DESC) ; une valeur invalide vaut ASC.
func ApplySortMapFor(q squirrel.SelectBuilder, sortMap SortMap, model any, tiebreak ...string) squirrel.SelectBuilder {
	t, err := structType(model)
	if err != nil {
		return ApplySortMap(q, sortMap)
	}
	known := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		if column := modelColumn(t.Field(i)); column != "" {
			known[column] = true
		}
	}

	fields := sortingFields(t)
	sort.SliceStable(fields, func(i, j int) bool {
		if fields[i].priority != fields[j].priority {
			return fields[i].priority < fields[j].priority
		}
		return fields[i].index < fields[j].index
	})

	var parts []string
	used := map[string]bool{}
	add := func(column, direction string) {
		if used[column] || !known[column] {
			return
		}
		used[column] = true
		parts = append(parts, column+" "+normalizeDirection(direction, "ASC"))
	}
	for _, f := range fields {
		if direction, ok := sortMap[f.column]; ok {
			add(f.column, direction)
		}
	}
	rest := make([]string, 0, len(sortMap))
	for column := range sortMap {
		if !used[column] {
			rest = append(rest, column)
		}
	}
	sort.Strings(rest)
	for _, column := range rest {
		add(column, sortMap[column])
	}
	for _, column := range tiebreak {
		add(strings.TrimSpace(column), "ASC")
	}
	if len(parts) == 0 {
		return q
	}
	return q.OrderBy(strings.Join(parts, ", "))
}

// camelToSnake convertit un nom CamelCase en snake_case (ex: Firstname -> firstname)
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
