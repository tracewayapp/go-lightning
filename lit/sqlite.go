package lit

import (
	"strings"
)

type SqliteInsertUpdateQueryGenerator struct{}

func (SqliteInsertUpdateQueryGenerator) GenerateInsertQuery(tableName string, columnKeys []string, hasIntId bool) (string, []string) {
	var insertQuery strings.Builder

	insertQuery.WriteString("INSERT INTO ")
	insertQuery.WriteString(sqliteEscapeReserved(tableName))
	insertQuery.WriteString(" (")

	totalKeys := len(columnKeys)
	for i, k := range columnKeys {
		insertQuery.WriteString(sqliteEscapeReserved(k))
		if i != totalKeys-1 {
			insertQuery.WriteString(",")
		}
	}

	insertQuery.WriteString(") VALUES (")

	insertColumns := []string{}
	for i, k := range columnKeys {
		if hasIntId && k == "id" {
			insertQuery.WriteString("NULL")
		} else {
			insertColumns = append(insertColumns, k)
			insertQuery.WriteString("?")
		}
		if i != totalKeys-1 {
			insertQuery.WriteString(",")
		}
	}
	insertQuery.WriteString(")")

	return insertQuery.String(), insertColumns
}

func (SqliteInsertUpdateQueryGenerator) GenerateUpdateQuery(tableName string, columnKeys []string) string {
	var updateQuery strings.Builder
	updateQuery.WriteString("UPDATE ")
	updateQuery.WriteString(sqliteEscapeReserved(tableName))
	updateQuery.WriteString(" SET ")

	totalKeys := len(columnKeys)
	for i, k := range columnKeys {
		updateQuery.WriteString(sqliteEscapeReserved(k))
		updateQuery.WriteString(" = ?")
		if i != totalKeys-1 {
			updateQuery.WriteString(",")
		}
	}

	updateQuery.WriteString(" WHERE ")

	return updateQuery.String()
}

func sqliteJoinStringForIn(count int) string {
	var sb strings.Builder
	for i := 0; i < count; i++ {
		sb.WriteString("?")
		if i < count-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func sqliteEscapeReserved(tableOrColumn string) string {
	escaped := strings.ReplaceAll(tableOrColumn, `"`, `""`)

	if _, exists := sqliteReservedKeywords[strings.ToUpper(tableOrColumn)]; exists {
		return `"` + escaped + `"`
	}
	return tableOrColumn
}

var sqliteReservedKeywords = map[string]struct{}{
	"ABORT":             {},
	"ACTION":            {},
	"ADD":               {},
	"AFTER":             {},
	"ALL":               {},
	"ALTER":             {},
	"ALWAYS":            {},
	"ANALYZE":           {},
	"AND":               {},
	"AS":                {},
	"ASC":               {},
	"ATTACH":            {},
	"AUTOINCREMENT":     {},
	"BEFORE":            {},
	"BEGIN":             {},
	"BETWEEN":           {},
	"BY":                {},
	"CASCADE":           {},
	"CASE":              {},
	"CAST":              {},
	"CHECK":             {},
	"COLLATE":           {},
	"COLUMN":            {},
	"COMMIT":            {},
	"CONFLICT":          {},
	"CONSTRAINT":        {},
	"CREATE":            {},
	"CROSS":             {},
	"CURRENT":           {},
	"CURRENT_DATE":      {},
	"CURRENT_TIME":      {},
	"CURRENT_TIMESTAMP": {},
	"DATABASE":          {},
	"DEFAULT":           {},
	"DEFERRABLE":        {},
	"DEFERRED":          {},
	"DELETE":            {},
	"DESC":              {},
	"DETACH":            {},
	"DISTINCT":          {},
	"DO":                {},
	"DROP":              {},
	"EACH":              {},
	"ELSE":              {},
	"END":               {},
	"ESCAPE":            {},
	"EXCEPT":            {},
	"EXCLUDE":           {},
	"EXCLUSIVE":         {},
	"EXISTS":            {},
	"EXPLAIN":           {},
	"FAIL":              {},
	"FILTER":            {},
	"FIRST":             {},
	"FOLLOWING":         {},
	"FOR":               {},
	"FOREIGN":           {},
	"FROM":              {},
	"FULL":              {},
	"GENERATED":         {},
	"GLOB":              {},
	"GROUP":             {},
	"GROUPS":            {},
	"HAVING":            {},
	"IF":                {},
	"IGNORE":            {},
	"IMMEDIATE":         {},
	"IN":                {},
	"INDEX":             {},
	"INDEXED":           {},
	"INITIALLY":         {},
	"INNER":             {},
	"INSERT":            {},
	"INSTEAD":           {},
	"INTERSECT":         {},
	"INTO":              {},
	"IS":                {},
	"ISNULL":            {},
	"JOIN":              {},
	"KEY":               {},
	"LAST":              {},
	"LEFT":              {},
	"LIKE":              {},
	"LIMIT":             {},
	"MATCH":             {},
	"MATERIALIZED":      {},
	"NATURAL":           {},
	"NO":                {},
	"NOT":               {},
	"NOTHING":           {},
	"NOTNULL":           {},
	"NULL":              {},
	"NULLS":             {},
	"OF":                {},
	"OFFSET":            {},
	"ON":                {},
	"OR":                {},
	"ORDER":             {},
	"OTHERS":            {},
	"OUTER":             {},
	"OVER":              {},
	"PARTITION":         {},
	"PLAN":              {},
	"PRAGMA":            {},
	"PRECEDING":         {},
	"PRIMARY":           {},
	"QUERY":             {},
	"RAISE":             {},
	"RANGE":             {},
	"RECURSIVE":         {},
	"REFERENCES":        {},
	"REGEXP":            {},
	"REINDEX":           {},
	"RELEASE":           {},
	"RENAME":            {},
	"REPLACE":           {},
	"RESTRICT":          {},
	"RETURNING":         {},
	"RIGHT":             {},
	"ROLLBACK":          {},
	"ROW":               {},
	"ROWS":              {},
	"SAVEPOINT":         {},
	"SELECT":            {},
	"SET":               {},
	"TABLE":             {},
	"TEMP":              {},
	"TEMPORARY":         {},
	"THEN":              {},
	"TIES":              {},
	"TO":                {},
	"TRANSACTION":       {},
	"TRIGGER":           {},
	"UNBOUNDED":         {},
	"UNION":             {},
	"UNIQUE":            {},
	"UPDATE":            {},
	"USING":             {},
	"VACUUM":            {},
	"VALUES":            {},
	"VIEW":              {},
	"VIRTUAL":           {},
	"WHEN":              {},
	"WHERE":             {},
	"WINDOW":            {},
	"WITH":              {},
	"WITHOUT":           {},
}
