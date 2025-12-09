package querybuilder

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Masterminds/squirrel"
)

type PaginationQuery struct {
	Page   uint64
	Limit  uint64
	Offset uint64
}

type PaginationQue struct {
	// --- Pagination & contrôle ---
	Page   int `validate:"gte=0"`
	Limit  int `validate:"gte=1,lte=1000"`
	Offset int `validate:"gte=0"`
}

func ApplyPagination(q squirrel.SelectBuilder, p PaginationQuery) squirrel.SelectBuilder {
	return q.Limit(p.Limit).Offset(p.Offset)
}

func Parse(fq *PaginationQuery, r *http.Request) (*PaginationQuery, error) {
	qs := r.URL.Query()

	// pagination
	if v := qs.Get("size"); v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			fq.Limit, err = SafeIntToUint64(l)
			if err != nil {
				return nil, err
			}
		}
	}
	if v := qs.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			fq.Page, err = SafeIntToUint64(p)
			if err != nil {
				return nil, err
			}
			fq.Offset = (fq.Page - 1) * fq.Limit
		}
	}
	return fq, nil
}

// SafeIntToUint64 converts an int to uint64, returning an error for negative values.
func SafeIntToUint64(i int) (uint64, error) {
	if i < 0 {
		return 0, errors.New("cannot convert negative int to uint64")
	}
	// Conversion is safe for zero and positive numbers
	return uint64(i), nil
}
