package models

import (
	"fmt"

	"github.com/iamgak/go-task/pkg"
)

type Filters struct {
	CurrPage  int
	PageSize  int
	Status    string
	SortBy    string
	SortOrder string
	DueAfter  string
	DueBefore string
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.CurrPage - 1) * f.PageSize
}

func (f Filters) sortDirection() string {
	if f.SortOrder == "asc" {
		return f.SortOrder
	}

	return "desc"
}

func (f *Filters) ValidStatus() bool {
	var v *pkg.Validator
	return v.ValidStatus(f.Status)
}

func (f Filters) sortColumn() string {
	sortSafeList := []string{"id", "due_at", "created_at", "updated_at"}
	for _, safeValue := range sortSafeList {
		if f.SortBy == safeValue {
			return f.SortBy
		}
	}

	return ""
}

func (f Filters) otherConditions() string {
	where := ""
	if f.ValidStatus() {
		where += fmt.Sprintf(" AND status = '%s'", f.Status)
	}

	if f.DueAfter != "" {
		where += fmt.Sprintf(" AND due_at >= '%s'", f.DueAfter)
	} else if f.DueBefore != "" {
		where += fmt.Sprintf(" AND due_at <= '%s'", f.DueAfter)
	}

	if f.sortColumn() != "" {
		where = fmt.Sprintf(" %s ORDER BY '%s' %s LIMIT %d OFFSET %d", where, f.SortBy, f.sortDirection(), f.limit(), f.offset())
	} else {
		where = fmt.Sprintf(" %s ORDER BY id DESC LIMIT %d OFFSET %d", where, f.limit(), f.offset())
	}

	return where
}
