package data

import (
	"strings"

	"github.com/martinezmoises/comments/internal/validator"
)

// Filters holds pagination and sorting information.
type Filters struct {
	Page         int      // Which page number to return.
	PageSize     int      // Number of records per page.
	Sort         string   // Sorting field (e.g., "id" or "-id").
	SortSafeList []string // Allowed fields for sorting.
}

// Metadata provides information about pagination.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// ValidateFilters ensures the Filters fields are valid.
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 500, "page", "must not exceed 500")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must not exceed 100")
	v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

// limit calculates the maximum number of records per page.
func (f Filters) limit() int {
	return f.PageSize
}

// offset calculates the number of records to skip for pagination.
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// calculateMetaData calculates pagination metadata.
func calculateMetaData(totalRecords, currentPage, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}

// sortColumn determines the column to sort by.
func (f Filters) sortColumn() string {
	if f.Sort == "" {
		return "id" // Default sort column.
	}
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	// Default to a safe column if sort is invalid.
	return "id"
}

// sortDirection determines the sorting direction (ASC or DESC).
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
