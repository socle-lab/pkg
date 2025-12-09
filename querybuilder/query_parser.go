package querybuilder

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// ParseQuery générique pour toutes les structs filtrables
func ParseQuery[T any](r *http.Request, fq T) (T, error) {
	qs := r.URL.Query()
	v := reflect.ValueOf(&fq).Elem()
	t := v.Type()

	// --- Pagination ---
	if field := v.FieldByName("Limit"); field.IsValid() && field.CanSet() {
		if val := qs.Get("size"); val != "" {
			if l, err := strconv.Atoi(val); err == nil {
				field.SetInt(int64(l))
			}
		}
	}

	if field := v.FieldByName("Page"); field.IsValid() && field.CanSet() {
		if val := qs.Get("page"); val != "" {
			if p, err := strconv.Atoi(val); err == nil {
				field.SetInt(int64(p))
				// Offset automatique
				if offsetField := v.FieldByName("Offset"); offsetField.IsValid() && offsetField.CanSet() {
					offsetField.SetInt(int64(p*int(v.FieldByName("Limit").Int()) - int(v.FieldByName("Limit").Int())))
				}
			}
		}
	}

	if field := v.FieldByName("LastID"); field.IsValid() && field.CanSet() {
		if val := qs.Get("last_id"); val != "" {
			if li, err := strconv.Atoi(val); err == nil {
				field.SetInt(int64(li))
			}
		}
	}

	// --- Champs filtrables ---
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() || field.Kind() != reflect.String {
			continue
		}

		tag := t.Field(i).Tag.Get("filter")
		if tag == "" {
			continue
		}

		// On prend juste le nom du paramètre avant la virgule
		param := strings.Split(tag, ",")[0]
		if val := qs.Get(param); val != "" {
			field.SetString(val)
		}
	}

	return fq, nil
}

// fq := models.VehicleRegistrationFilteredQuery{}
// fq, err := utils.ParseQuery(r, fq)
// if err != nil {
//     // gérer l'erreur
// }
