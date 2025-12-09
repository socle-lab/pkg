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
