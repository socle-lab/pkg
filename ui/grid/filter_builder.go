package grid

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// BuildFilterFieldsFromModel génère automatiquement les champs de filtre à partir des tags filter d'un modèle
// Retourne une slice de GridFilterField prête à être utilisée dans un GridFilter
func BuildFilterFieldsFromModel(modelStruct interface{}) []GridFilterField {
	var fields []GridFilterField

	// Utiliser la réflexion pour parcourir les champs de la struct
	v := reflect.ValueOf(modelStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fields
	}

	t := v.Type()

	// Parcourir tous les champs de la struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("filter")
		if tag == "" || tag == "-" {
			continue
		}

		// Parser le tag: "user_login,criteria=ILIKE"
		parts := strings.Split(tag, ",")
		if len(parts) == 0 {
			continue
		}

		// La première partie est le nom du champ (clé pour le query parameter)
		fieldNameKey := strings.TrimSpace(parts[0])
		if fieldNameKey == "" {
			continue
		}

		// Extraire le critère par défaut du tag
		defaultCriteria := ""
		for j := 1; j < len(parts); j++ {
			part := strings.TrimSpace(parts[j])
			if strings.HasPrefix(part, "criteria=") {
				defaultCriteria = strings.ToUpper(strings.TrimPrefix(part, "criteria="))
				break
			} else if strings.HasPrefix(part, "type=") {
				// Support de compatibilité pour l'ancien format
				defaultCriteria = strings.ToUpper(strings.TrimPrefix(part, "type="))
				break
			}
		}

		// Le nom du query parameter (préfixé avec "filter_")
		queryParamName := fmt.Sprintf("filter_%s", fieldNameKey)

		// Générer un label à partir du nom du champ Go
		label := generateLabel(field.Name)

		// Déterminer le type de filtre selon le critère et le type Go
		filterType := determineFilterType(field.Type, defaultCriteria)

		// Créer le champ de filtre
		filterField := NewGridFilterField(queryParamName, label, filterType)
		fields = append(fields, filterField)
	}

	return fields
}

// generateLabel convertit un nom de champ Go en label lisible
// Ex: UserLogin -> "User Login", CreatedAt -> "Created At"
func generateLabel(fieldName string) string {
	var result []rune
	for i, r := range fieldName {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, ' ')
		}
		result = append(result, r)
	}
	return string(result)
}

// determineFilterType détermine le type de filtre selon le type Go et le critère
func determineFilterType(fieldType reflect.Type, criteria string) FilterType {
	// Détecter les champs time.Time pour les filtres de date
	if fieldType.String() == "time.Time" {
		return FilterText
	}

	// Si le critère est BETWEEN, utiliser un champ texte (l'utilisateur entrera "min-max")
	if criteria == "BETWEEN" {
		return FilterText
	}

	// Si le critère est IN, utiliser un champ texte (l'utilisateur entrera "val1,val2,val3")
	if criteria == "IN" {
		return FilterText
	}

	// Selon le type Go du champ
	switch fieldType.Kind() {
	case reflect.Bool:
		return FilterBoolean
	case reflect.String:
		// Pour les critères LIKE/ILIKE, utiliser FilterText
		if criteria == "LIKE" || criteria == "ILIKE" || criteria == "" {
			return FilterText
		}
		return FilterText
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return FilterText
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return FilterText
	case reflect.Float32, reflect.Float64:
		return FilterText
	default:
		return FilterText
	}
}
