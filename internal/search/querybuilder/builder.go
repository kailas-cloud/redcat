package querybuilder

import (
	"strconv"
	"strings"
)

// CategoriesOR builds a TAG filter like @category_ids:{a|b|c} or returns "*" when empty.
func CategoriesOR(field string, ids []string) string {
	clean := make([]string, 0, len(ids))
	for _, v := range ids {
		v = strings.TrimSpace(v)
		if v != "" { clean = append(clean, v) }
	}
	if len(clean) == 0 { return "*" }
	return "@" + field + ":{" + strings.Join(clean, "|") + "}"
}

// KNN appends the KNN clause to the filter.
func KNN(filter string, k int64, vectorField string) string {
	if strings.TrimSpace(filter) == "" { filter = "*" }
	return filter + "=>[KNN " + itoa(k) + " @" + vectorField + " $vec]"
}

func itoa(k int64) string { return strconv.FormatInt(k, 10) }
