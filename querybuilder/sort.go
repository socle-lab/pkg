package querybuilder

import (
	"fmt"
	"reflect"
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
