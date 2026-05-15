package querybuilder

import (
	"net/http"

	"github.com/socle-lab/pkg/validator"
)

// ParseCollectionQuery parses pagination, filters, and sort from the request,
// then validates the pagination struct. Returns (nil, nil, nil, err) on any failure.
func ParseCollectionQuery(r *http.Request, structType any, limit uint64) (*PaginationQuery, FilterMap, SortMap, error) {
	paginationQuery := &PaginationQuery{
		Page:   1,
		Limit:  limit,
		Offset: 0,
	}
	var err error
	paginationQuery, err = ParsePaginationQuery(paginationQuery, r)
	if err != nil {
		return nil, nil, nil, err
	}
	filterMap, err := ParseFilterMap(structType, r)
	if err != nil {
		return nil, nil, nil, err
	}
	sortMap, err := ParseSortMap(structType, r)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := validator.Validate.Struct(paginationQuery); err != nil {
		return nil, nil, nil, err
	}
	return paginationQuery, filterMap, sortMap, nil
}
