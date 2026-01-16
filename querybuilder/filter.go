package querybuilder

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
)

type FilterField struct {
	Column string
	Value  any
	Op     string // default: "ILIKE", can be "=", "IN", etc.
}

func ApplyFilters(q squirrel.SelectBuilder, filters []FilterField) squirrel.SelectBuilder {
	for _, f := range filters {
		if f.Value == nil {
			continue
		}
		op := f.Op
		if op == "" {
			op = "="
		}
		clause := fmt.Sprintf("%s %s ?", f.Column, op)
		q = q.Where(clause, f.Value)
	}
	return q
}

// BuildWhereFromStruct parcourt une struct de filtre et construit une clause WHERE dynamique
func ApplyFiltersFromStruct(q squirrel.SelectBuilder, filter any) (squirrel.SelectBuilder, error) {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return q, fmt.Errorf("filter must be a struct")
	}

	t := v.Type()
	var conditions []FilterField
	argIndex := 1

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		tag := field.Tag.Get("filter")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse tag: "column_name,type=ilike"
		parts := strings.Split(tag, ",")
		if len(parts) == 0 {
			continue
		}

		column := strings.TrimSpace(strings.TrimPrefix(parts[0], "filter:"))
		if column == "-" || column == "" {
			continue
		}

		var filterType string
		if len(parts) > 1 && strings.HasPrefix(parts[1], "type=") {
			filterType = strings.TrimPrefix(parts[1], "type=")
		} else {
			filterType = "ILIKE" // default fallback
		}

		// Skip zero values
		if isZeroValue(value) {
			continue
		}

		// Build condition
		switch filterType {
		case "ILIKE":
			conditions = append(conditions, FilterField{
				Column: column,
				Value:  fmt.Sprintf("%%%v%%", value.Interface()),
				Op:     filterType,
			})

		case "LIKE":
			conditions = append(conditions, FilterField{
				Column: column,
				Value:  fmt.Sprintf("%%%v%%", value.Interface()),
				Op:     filterType,
			})

		default:
			conditions = append(conditions, FilterField{
				Column: column,
				Value:  value.Interface(),
				Op:     filterType,
			})
		}

		argIndex++
	}

	q = ApplyFilters(q, conditions)

	return q, nil
}

// BuildWhereFromStruct parcourt une struct de filtre et construit une clause WHERE dynamique
func BuildWhereFromStruct(filter any) (string, []any, error) {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("filter must be a struct")
	}

	t := v.Type()
	var conditions []string
	var args []any
	argIndex := 1

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		tag := field.Tag.Get("filter")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse tag: "column_name,type=ilike"
		parts := strings.Split(tag, ",")
		if len(parts) == 0 {
			continue
		}

		column := strings.TrimSpace(strings.TrimPrefix(parts[0], "filter:"))
		if column == "-" || column == "" {
			continue
		}

		var filterType string
		if len(parts) > 1 && strings.HasPrefix(parts[1], "type=") {
			filterType = strings.TrimPrefix(parts[1], "type=")
		} else {
			filterType = "eq" // default fallback
		}

		// Skip zero values
		if isZeroValue(value) {
			continue
		}

		// Build condition
		switch filterType {
		case "eq":
			conditions = append(conditions, fmt.Sprintf("%s = $%d", column, argIndex))
			args = append(args, value.Interface())

		case "gt":
			conditions = append(conditions, fmt.Sprintf("%s > $%d", column, argIndex))
			args = append(args, value.Interface())

		case "lt":
			conditions = append(conditions, fmt.Sprintf("%s < $%d", column, argIndex))
			args = append(args, value.Interface())

		case "ilike":
			conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", column, argIndex))
			args = append(args, fmt.Sprintf("%%%v%%", value.Interface()))

		case "like":
			conditions = append(conditions, fmt.Sprintf("%s LIKE $%d", column, argIndex))
			args = append(args, fmt.Sprintf("%%%v%%", value.Interface()))

		default:
			return "", nil, fmt.Errorf("unsupported filter type: %s", filterType)
		}

		argIndex++
	}

	if len(conditions) == 0 {
		return "", nil, nil
	}

	whereSQL := "WHERE " + strings.Join(conditions, " AND ")
	return whereSQL, args, nil
}

func ParseFilter(fq any, r *http.Request) (any, error) {
	qs := r.URL.Query()

	// dynamic fill from filter tags
	val := reflect.ValueOf(fq).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("filter")
		if tag == "" {
			continue
		}
		// tag example: "application_id,type=ILIKE" or "application_id,type="
		parts := strings.Split(tag, ",")
		col := strings.TrimSpace(parts[0])
		if col == "" {
			continue
		}
		if v := qs.Get(col); v != "" {
			f := val.Field(i)
			if !f.CanSet() {
				continue
			}
			switch f.Kind() {
			case reflect.String:
				f.SetString(v)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				// si tu poses des champs int, tu peux parser ici ; pour l'instant nos filtres sont string
				if n, err := strconv.Atoi(v); err == nil {
					f.SetInt(int64(n))
				}
			// ajouter d'autres cas si nécessaire
			default:
				// ignore unsupported kinds for now
			}
		}
	}

	return fq, nil
}

// isZeroValue vérifie si une valeur est vide / zéro
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		// Compare à la valeur zéro par défaut
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

// FilterMap représente la structure d'un filtre dans la map retournée
type FilterMap map[string]map[string]interface{}

// ParseFilterMap parse les query parameters en se basant sur les tags filter d'une struct modèle
// et retourne une map structurée où :
// - La clé principale est le nom du champ (ex: "user_login")
// - La valeur est une map avec "value" et "criteria"
// Le critère peut être surchargé via un query parameter "{key}_criteria"
func ParseFilterMap(modelStruct interface{}, r *http.Request) (FilterMap, error) {
	qs := r.URL.Query()
	filterMap := make(FilterMap)

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
		tag := field.Tag.Get("filter")
		if tag == "" || tag == "-" {
			continue
		}

		// Parser le tag: "user_login,criteria=ILIKE" ou "user_login,type=ILIKE" (compatibilité)
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
		// Le nom du query parameter (préfixé avec "filter_")
		queryParamName := fmt.Sprintf("filter_%s", fieldNameKey)

		// Extraire le critère par défaut du tag
		defaultCriteria := ""
		for j := 1; j < len(parts); j++ {
			part := strings.TrimSpace(parts[j])
			if strings.HasPrefix(part, "criteria=") {
				defaultCriteria = strings.TrimPrefix(part, "criteria=")
				break
			} else if strings.HasPrefix(part, "type=") {
				// Support de compatibilité pour l'ancien format
				defaultCriteria = strings.TrimPrefix(part, "type=")
				break
			}
		}

		// Vérifier si le query parameter existe pour ce champ
		value := qs.Get(queryParamName)
		if value == "" {
			continue
		}

		// Vérifier si le critère est surchargé via un query parameter
		criteria := defaultCriteria
		if criteriaOverride := qs.Get(queryParamName + "_criteria"); criteriaOverride != "" {
			criteria = strings.ToUpper(criteriaOverride)
		}

		// Parser la valeur selon le type de critère
		filterData := make(map[string]interface{})
		filterData["criteria"] = criteria

		switch criteria {
		case "DATE":
			// Pour DATE, on attend une date au format dd/MM/yyyy
			filterData["value"] = strings.TrimSpace(value)

		case "BETWEEN":
			// Pour BETWEEN, on attend deux valeurs séparées par un tiret (-)
			parts := strings.Split(value, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("BETWEEN filter requires two values separated by '-' for field %s", queryParamName)
			}
			filterData["value"] = []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}

		case "IN":
			// Pour IN, on attend plusieurs valeurs séparées par des virgules
			parts := strings.Split(value, ",")
			values := make([]string, 0, len(parts))
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					values = append(values, trimmed)
				}
			}
			if len(values) == 0 {
				continue // Ignorer si aucune valeur valide
			}
			filterData["value"] = values

		case "LIKE", "ILIKE":
			// Pour LIKE et ILIKE, la valeur est une string simple
			filterData["value"] = value

		case "=", "":
			// Pour l'égalité (ou critère vide), la valeur est une string simple
			filterData["value"] = value
			if criteria == "" {
				filterData["criteria"] = "="
			}

		default:
			// Par défaut, traiter comme une string simple
			filterData["value"] = value
		}

		// Utiliser le nom de la colonne DB comme clé dans la map (pas le query parameter)
		filterMap[dbColumnName] = filterData
	}

	return filterMap, nil
}

// ApplyFilterMap applique les filtres d'une FilterMap à une requête squirrel.SelectBuilder
func ApplyFilterMap(q squirrel.SelectBuilder, filterMap FilterMap) (squirrel.SelectBuilder, error) {
	for fieldName, filterData := range filterMap {
		value, ok := filterData["value"]
		if !ok || value == nil {
			continue
		}

		criteria, ok := filterData["criteria"].(string)
		if !ok {
			criteria = "="
		}

		switch criteria {
		case "DATE":
			// Pour DATE, accepter dd/MM/yyyy ou yyyy-MM-dd et filtrer par date exacte
			strValue, ok := value.(string)
			if !ok {
				continue
			}

			var sqlDate string
			// Détecter le format de la date
			if strings.Contains(strValue, "/") {
				// Format dd/MM/yyyy -> convertir en yyyy-MM-dd
				dateParts := strings.Split(strValue, "/")
				if len(dateParts) == 3 {
					sqlDate = fmt.Sprintf("%s-%s-%s", dateParts[2], dateParts[1], dateParts[0])
				}
			} else if strings.Contains(strValue, "-") {
				// Format yyyy-MM-dd -> utiliser directement
				sqlDate = strValue
			}

			// Filtrer par date exacte si le format est valide
			if sqlDate != "" {
				q = q.Where(squirrel.Expr(fmt.Sprintf("DATE(%s) = ?", fieldName), sqlDate))
			}

		case "ILIKE", "LIKE":
			// Pour LIKE et ILIKE, la valeur doit être une string
			strValue, ok := value.(string)
			if !ok {
				continue
			}
			// Échapper les caractères spéciaux SQL (_ et %) pour éviter qu'ils soient interprétés comme des wildcards
			// Le _ correspond à un caractère, le % à plusieurs caractères
			escapedValue := escapeLikePattern(strValue)
			// Ajouter les wildcards pour la recherche partielle
			pattern := fmt.Sprintf("%%%s%%", escapedValue)
			if criteria == "ILIKE" {
				q = q.Where(squirrel.Expr(fmt.Sprintf("%s ILIKE ?", fieldName), pattern))
			} else {
				q = q.Where(squirrel.Expr(fmt.Sprintf("%s LIKE ?", fieldName), pattern))
			}

		case "BETWEEN":
			// Pour BETWEEN, la valeur doit être un slice de deux strings
			values, ok := value.([]string)
			if !ok || len(values) != 2 {
				continue
			}
			q = q.Where(squirrel.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", fieldName), values[0], values[1]))

		case "IN":
			// Pour IN, la valeur doit être un slice de strings
			values, ok := value.([]string)
			if !ok || len(values) == 0 {
				continue
			}
			// Convertir le slice en interface{} pour squirrel
			args := make([]interface{}, len(values))
			for i, v := range values {
				args[i] = v
			}
			q = q.Where(squirrel.Eq{fieldName: args})

		case "=", "":
			// Pour l'égalité, la valeur est une string simple
			strValue, ok := value.(string)
			if !ok {
				continue
			}
			q = q.Where(squirrel.Eq{fieldName: strValue})

		default:
			// Pour les autres critères, traiter comme une égalité
			strValue, ok := value.(string)
			if !ok {
				continue
			}
			q = q.Where(squirrel.Expr(fmt.Sprintf("%s %s ?", fieldName, criteria), strValue))
		}
	}

	return q, nil
}

// escapeLikePattern échappe les caractères spéciaux SQL (_ et %) dans les patterns LIKE/ILIKE
// En SQL, _ correspond à un caractère et % à plusieurs caractères
// Pour chercher ces caractères littéralement, il faut les échapper avec un backslash
func escapeLikePattern(s string) string {
	// Remplacer \ par \\ d'abord (pour éviter d'échapper les échappements)
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Puis échapper _ et %
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "%", "\\%")
	return s
}
