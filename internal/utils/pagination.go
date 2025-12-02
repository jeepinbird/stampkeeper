package utils

import (
	"math"

	"github.com/jeepinbird/stampkeeper/internal/models"
)

// CalculatePagination creates pagination data from total count, current page, and items per page
func CalculatePagination(totalItems int64, page, limit int) models.Pagination {
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return models.Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
		NextPage:    page + 1,
		PrevPage:    page - 1,
	}
}
