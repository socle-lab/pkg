package querybuilder

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Masterminds/squirrel"
)

type PaginationQuery struct {
	Page   uint64 `validate:"gte=1"`
	Limit  uint64 `validate:"gte=1"`
	Offset uint64 `validate:"gte=0"`
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

func ParsePaginationQuery(fq *PaginationQuery, r *http.Request) (*PaginationQuery, error) {
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

// BuildPaginationLinks returns current, previous, and next links for a paginated request.
// currentItemsCount is the number of items returned for the current page.
// totalItemsCount is the total number of items across all pages.
func BuildPaginationLinks(r *http.Request, pq PaginationQuery, currentItemsCount, totalItemsCount int) (current, previous, next string) {
	if r == nil {
		return "", "", ""
	}

	current = r.RequestURI

	q := r.URL.Query()
	q.Del("page")
	q.Del("size")
	basePath := r.URL.Path
	queryString := q.Encode()

	if pq.Page > 1 {
		previous = fmt.Sprintf("%s?page=%d&size=%d", basePath, pq.Page-1, pq.Limit)
		if queryString != "" {
			previous += "&" + queryString
		}
	}

	if int(pq.Offset)+currentItemsCount < totalItemsCount {
		next = fmt.Sprintf("%s?page=%d&size=%d", basePath, pq.Page+1, pq.Limit)
		if queryString != "" {
			next += "&" + queryString
		}
	}

	return current, previous, next
}

// SafeIntToUint64 converts an int to uint64, returning an error for negative values.
func SafeIntToUint64(i int) (uint64, error) {
	if i < 0 {
		return 0, errors.New("cannot convert negative int to uint64")
	}
	// Conversion is safe for zero and positive numbers
	return uint64(i), nil
}
