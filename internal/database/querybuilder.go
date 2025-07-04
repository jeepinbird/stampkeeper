package database

import (
	"fmt"
	"strings"
)

// QueryBuilder helps construct PostgreSQL queries with automatic parameter numbering
type QueryBuilder struct {
	query string
	args  []interface{}
}

// NewQueryBuilder creates a new QueryBuilder with a base query
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		query: baseQuery,
		args:  []interface{}{},
	}
}

// AddParam adds a parameter to the query and returns the PostgreSQL placeholder ($1, $2, etc.)
func (qb *QueryBuilder) AddParam(value interface{}) string {
	qb.args = append(qb.args, value)
	return fmt.Sprintf("$%d", len(qb.args))
}

// AddCondition adds a SQL condition to the query, replacing ? placeholders with proper PostgreSQL parameters
func (qb *QueryBuilder) AddCondition(condition string, values ...interface{}) {
	if len(values) == 0 {
		qb.query += condition
		return
	}
	
	// Replace ? placeholders with actual parameter numbers
	for _, value := range values {
		condition = strings.Replace(condition, "?", qb.AddParam(value), 1)
	}
	qb.query += condition
}

// GetQuery returns the final query string and parameters
func (qb *QueryBuilder) GetQuery() (string, []interface{}) {
	return qb.query, qb.args
}

// AddSearchFilter adds search conditions for name, scott_number, and series columns
func (qb *QueryBuilder) AddSearchFilter(searchTerm string, tableAlias string) {
	if searchTerm != "" {
		searchParam := "%" + searchTerm + "%"
		qb.AddCondition(fmt.Sprintf(` AND (LOWER(%s.name) LIKE LOWER(?) OR LOWER(%s.scott_number) LIKE LOWER(?) OR LOWER(%s.series) LIKE LOWER(?))`,
			tableAlias, tableAlias, tableAlias),
			searchParam, searchParam, searchParam)
	}
}

// AddBoxFilter adds a condition to filter by box_id
func (qb *QueryBuilder) AddBoxFilter(boxID string, instanceAlias string) {
	if boxID != "" {
		qb.AddCondition(fmt.Sprintf(` AND %s.box_id = ?`, instanceAlias), boxID)
	}
}

// AddOwnedFilter adds HAVING clause for owned/not owned stamps
func (qb *QueryBuilder) AddOwnedFilter(owned string, instanceAlias string) {
	if owned == "true" {
		qb.AddCondition(fmt.Sprintf(` HAVING COUNT(%s.id) > 0`, instanceAlias))
	} else if owned == "false" {
		qb.AddCondition(fmt.Sprintf(` HAVING COUNT(%s.id) = 0`, instanceAlias))
	}
}

// AddSortAndLimit adds ORDER BY, LIMIT, and OFFSET clauses
func (qb *QueryBuilder) AddSortAndLimit(sort, order string, limit, offset int, tableAlias string) {
	orderDir := "ASC"
	if strings.ToUpper(order) == "DESC" {
		orderDir = "DESC"
	}

	// Add a secondary sort by ID to ensure a stable, deterministic order for pagination.
	secondarySort := fmt.Sprintf(", %s.id ASC", tableAlias)
	
	switch sort {
	case "name":
		qb.AddCondition(fmt.Sprintf(` ORDER BY %s.name %s%s`, tableAlias, orderDir, secondarySort))
	case "issue_date":
		qb.AddCondition(fmt.Sprintf(` ORDER BY %s.issue_date %s%s`, tableAlias, orderDir, secondarySort))
	case "date_added":
		qb.AddCondition(fmt.Sprintf(` ORDER BY %s.date_added %s%s`, tableAlias, orderDir, secondarySort))
	default:
		qb.AddCondition(fmt.Sprintf(` 
			ORDER BY CASE WHEN %s.scott_number ~ '^\d+' THEN 
							CAST(SUBSTRING(%s.scott_number FROM '\d+') AS INTEGER) 
						ELSE 999999 
					 END %s,
					 %s.scott_number %s%s`, 
					tableAlias, tableAlias, orderDir, tableAlias, orderDir, secondarySort))
	}
	
	qb.AddCondition(` LIMIT ? OFFSET ?`, limit, offset)
}

// AddWhereCondition adds a generic WHERE condition with table/column and operator
func (qb *QueryBuilder) AddWhereCondition(column string, operator string, value interface{}) {
	qb.AddCondition(fmt.Sprintf(` AND %s %s ?`, column, operator), value)
}

// AddDeletedFilter adds a condition to filter out soft-deleted records
func (qb *QueryBuilder) AddDeletedFilter(tableAlias string) {
	qb.AddCondition(fmt.Sprintf(` AND %s.date_deleted IS NULL`, tableAlias))
}

// AddJumpToFilter adds a condition to filter stamps with Scott numbers >= specified value
func (qb *QueryBuilder) AddJumpToFilter(jumpToScottNumber string, tableAlias string) {
	if jumpToScottNumber != "" {
		// For numeric Scott numbers, compare numerically
		// For non-numeric Scott numbers, compare alphabetically
		qb.AddCondition(fmt.Sprintf(`
			AND (
				(%s.scott_number ~ '^\d+' AND CAST(SUBSTRING(%s.scott_number FROM '\d+') AS INTEGER) >= ?)
				OR
				(%s.scott_number !~ '^\d+' AND %s.scott_number >= ?)
			)`, tableAlias, tableAlias, tableAlias, tableAlias),
			jumpToScottNumber, jumpToScottNumber)
	}
}