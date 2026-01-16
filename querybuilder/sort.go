package querybuilder

import (
	"fmt"
	"net/http"
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

// SortMap représente la structure d'un tri dans la map retournée
type SortMap map[string]string // clé: nom de colonne DB, valeur: direction (ASC/DESC)

// ParseSortMap parse les query parameters en se basant sur les tags sorting d'une struct modèle
// et retourne une map structurée où :
// - La clé principale est le nom de la colonne DB (ex: "user_login")
// - La valeur est la direction du tri (ASC ou DESC)
// La direction peut être surchargée via un query parameter "sort_{key}_order"
func ParseSortMap(modelStruct interface{}, r *http.Request) (SortMap, error) {
	qs := r.URL.Query()
	sortMap := make(SortMap)

	// Utiliser la réflexion pour parcourir les champs de la struct
	v := reflect.ValueOf(modelStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("modelStruct must be a struct or pointer to struct")
	}

	t := v.Type()

	// Parcourir tous les champs de la struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("sorting")
		if tag == "" || tag == "-" {
			continue
		}

		// Parser le tag: "user_login,order=ASC" ou "user_login,order=DESC"
		parts := strings.Split(tag, ",")
		if len(parts) == 0 {
			continue
		}

		// La première partie est le nom du champ (clé pour le query parameter)
		fieldNameKey := strings.TrimSpace(parts[0])
		if fieldNameKey == "" {
			continue
		}

		// Le nom de la colonne DB (utilisé dans la map et dans les requêtes SQL) - sécurité : utiliser le tag db
		dbColumnName := field.Tag.Get("db")
		if dbColumnName == "" || dbColumnName == "-" {
			// Si pas de tag db, convertir le nom du champ Go en snake_case (ex: UserLogin -> user_login)
			dbColumnName = camelToSnake(field.Name)
		}

		// Extraire l'ordre par défaut du tag
		defaultOrder := "ASC"
		for j := 1; j < len(parts); j++ {
			part := strings.TrimSpace(parts[j])
			if strings.HasPrefix(part, "order=") {
				defaultOrder = strings.ToUpper(strings.TrimPrefix(part, "order="))
				break
			}
		}

		// Le nom du query parameter (préfixé avec "sorting_")
		queryParamName := fmt.Sprintf("sorting_%s", fieldNameKey)

		// Le tri est activé automatiquement si le tag sorting existe
		// L'ordre peut être surchargé via un query parameter
		order := defaultOrder
		if orderOverride := qs.Get(queryParamName + "_order"); orderOverride != "" {
			order = strings.ToUpper(orderOverride)
		}

		// Valider l'ordre (doit être ASC ou DESC)
		if order != "ASC" && order != "DESC" {
			order = "ASC" // Valeur par défaut si invalide
		}

		sortMap[dbColumnName] = order
	}

	return sortMap, nil
}

// ApplySortMap applique les tris d'une SortMap à une requête squirrel.SelectBuilder
func ApplySortMap(q squirrel.SelectBuilder, sortMap SortMap) squirrel.SelectBuilder {
	if sortMap == nil || len(sortMap) == 0 {
		return q
	}

	// Construire la clause ORDER BY
	var orderByParts []string
	for column, direction := range sortMap {
		orderByParts = append(orderByParts, fmt.Sprintf("%s %s", column, direction))
	}

	if len(orderByParts) > 0 {
		q = q.OrderBy(strings.Join(orderByParts, ", "))
	}

	return q
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
