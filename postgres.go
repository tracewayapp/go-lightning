package lit

import (
	"fmt"
	"strconv"
	"strings"
)

type pgDriver struct{}

var PostgreSQL Driver = &pgDriver{}

func (d *pgDriver) Name() string { return "PostgreSQL" }

func (d *pgDriver) String() string { return d.Name() }

func (d *pgDriver) GenerateInsertQuery(tableName string, columnKeys []string, hasIntId bool) (string, []string) {
	var insertQuery strings.Builder

	insertQuery.WriteString("INSERT INTO ")
	insertQuery.WriteString(pgEscapeReserved(tableName))
	insertQuery.WriteString(" (")

	totalKeys := len(columnKeys)
	for i, k := range columnKeys {
		insertQuery.WriteString(pgEscapeReserved(k))
		if i != totalKeys-1 {
			insertQuery.WriteString(",")
		}
	}

	insertQuery.WriteString(") VALUES (")

	counter := 1
	insertColumns := []string{}
	for i, k := range columnKeys {
		if hasIntId && k == "id" {
			insertQuery.WriteString("DEFAULT")
		} else {
			insertColumns = append(insertColumns, k)
			insertQuery.WriteString("$" + strconv.Itoa(counter))
			counter++
		}
		if i != totalKeys-1 {
			insertQuery.WriteString(",")
		}
	}
	insertQuery.WriteString(") RETURNING id")

	return insertQuery.String(), insertColumns
}

func (d *pgDriver) GenerateUpdateQuery(tableName string, columnKeys []string) string {
	var updateQuery strings.Builder
	updateQuery.WriteString("UPDATE ")
	updateQuery.WriteString(pgEscapeReserved(tableName))
	updateQuery.WriteString(" SET ")

	totalKeys := len(columnKeys)
	for i, k := range columnKeys {
		updateQuery.WriteString(pgEscapeReserved(k))
		updateQuery.WriteString(" = $" + strconv.Itoa(i+1))
		if i != totalKeys-1 {
			updateQuery.WriteString(",")
		}
	}

	updateQuery.WriteString(" WHERE ")

	return updateQuery.String()
}

func (d *pgDriver) InsertAndGetId(ex Executor, query string, args ...any) (int, error) {
	row := ex.QueryRow(query, args...)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *pgDriver) Placeholder(argIndex int) string {
	return "$" + strconv.Itoa(argIndex)
}

func (d *pgDriver) SupportsBackslashEscape() bool { return false }

func (d *pgDriver) RenumberWhereClause(where string, offset int) string {
	return pgRenumberPlaceholders(where, offset)
}

func (d *pgDriver) JoinStringForIn(offset int, count int) string {
	return pgJoinStringForIn(offset, count)
}

// Deprecated: Use PostgreSQL variable directly. PgInsertUpdateQueryGenerator is kept for backward compatibility.
type PgInsertUpdateQueryGenerator = pgDriver

func pgRenumberPlaceholders(where string, offset int) string {
	if !strings.Contains(where, "$") {
		return where
	}

	var newWhere strings.Builder
	parsingIdentifier := false

	for _, c := range where {
		if c == '$' {
			parsingIdentifier = true
			newWhere.WriteRune(c)
		} else if parsingIdentifier {
			if c >= '0' && c <= '9' {
				continue
			} else {
				parsingIdentifier = false
				offset++
				newWhere.WriteString(strconv.Itoa(offset))
				newWhere.WriteRune(c)
			}
		} else {
			newWhere.WriteRune(c)
		}
	}
	if parsingIdentifier {
		offset++
		newWhere.WriteString(strconv.Itoa(offset))
	}

	return newWhere.String()
}

func pgJoinStringForIn(offset int, count int) string {
	var sb strings.Builder
	for i := 0; i < count; i++ {
		sb.WriteString("$" + strconv.Itoa(i+1+offset))
		if i < count-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func pgEscapeReserved(tableOrColumn string) string {
	escaped := strings.ReplaceAll(tableOrColumn, `"`, `""`)

	if _, exists := pgReservedKeywords[strings.ToUpper(tableOrColumn)]; exists {
		return `"` + escaped + `"`
	}
	return tableOrColumn
}

// ensure pgDriver implements Driver at compile time
var _ Driver = (*pgDriver)(nil)
var _ fmt.Stringer = (*pgDriver)(nil)

var pgReservedKeywords = map[string]struct{}{
	"ABORT":             {},
	"ABSENT":            {},
	"ABSOLUTE":          {},
	"ACCESS":            {},
	"ACTION":            {},
	"ADD":               {},
	"ADMIN":             {},
	"AFTER":             {},
	"AGGREGATE":         {},
	"ALL":               {},
	"ALSO":              {},
	"ALTER":             {},
	"ALWAYS":            {},
	"ANALYSE":           {},
	"ANALYZE":           {},
	"AND":               {},
	"ANY":               {},
	"ARRAY":             {},
	"AS":                {},
	"ASC":               {},
	"ASENSITIVE":        {},
	"ASSERTION":         {},
	"ASSIGNMENT":        {},
	"ASYMMETRIC":        {},
	"AT":                {},
	"ATOMIC":            {},
	"ATTACH":            {},
	"ATTRIBUTE":         {},
	"AUTHORIZATION":     {},
	"BACKWARD":          {},
	"BEFORE":            {},
	"BEGIN":             {},
	"BETWEEN":           {},
	"BIGINT":            {},
	"BINARY":            {},
	"BIT":               {},
	"BOOLEAN":           {},
	"BOTH":              {},
	"BREADTH":           {},
	"BY":                {},
	"CACHE":             {},
	"CALL":              {},
	"CALLED":            {},
	"CASCADE":           {},
	"CASCADED":          {},
	"CASE":              {},
	"CAST":              {},
	"CATALOG":           {},
	"CHAIN":             {},
	"CHAR":              {},
	"CHARACTER":         {},
	"CHARACTERISTICS":   {},
	"CHECK":             {},
	"CHECKPOINT":        {},
	"CLASS":             {},
	"CLOSE":             {},
	"CLUSTER":           {},
	"COALESCE":          {},
	"COLLATE":           {},
	"COLLATION":         {},
	"COLUMN":            {},
	"COLUMNS":           {},
	"COMMENT":           {},
	"COMMENTS":          {},
	"COMMIT":            {},
	"COMMITTED":         {},
	"COMPRESSION":       {},
	"CONCURRENTLY":      {},
	"CONDITIONAL":       {},
	"CONFIGURATION":     {},
	"CONFLICT":          {},
	"CONNECTION":        {},
	"CONSTRAINT":        {},
	"CONSTRAINTS":       {},
	"CONTENT":           {},
	"CONTINUE":          {},
	"CONVERSION":        {},
	"COPY":              {},
	"COST":              {},
	"CREATE":            {},
	"CROSS":             {},
	"CSV":               {},
	"CUBE":              {},
	"CURRENT":           {},
	"CURRENT_CATALOG":   {},
	"CURRENT_DATE":      {},
	"CURRENT_ROLE":      {},
	"CURRENT_SCHEMA":    {},
	"CURRENT_TIME":      {},
	"CURRENT_TIMESTAMP": {},
	"CURRENT_USER":      {},
	"CURSOR":            {},
	"CYCLE":             {},
	"DATA":              {},
	"DATABASE":          {},
	"DAY":               {},
	"DEALLOCATE":        {},
	"DEC":               {},
	"DECIMAL":           {},
	"DECLARE":           {},
	"DEFAULT":           {},
	"DEFAULTS":          {},
	"DEFERRABLE":        {},
	"DEFERRED":          {},
	"DEFINER":           {},
	"DELETE":            {},
	"DELIMITER":         {},
	"DELIMITERS":        {},
	"DEPENDS":           {},
	"DEPTH":             {},
	"DESC":              {},
	"DETACH":            {},
	"DICTIONARY":        {},
	"DISABLE":           {},
	"DISCARD":           {},
	"DISTINCT":          {},
	"DO":                {},
	"DOCUMENT":          {},
	"DOMAIN":            {},
	"DOUBLE":            {},
	"DROP":              {},
	"EACH":              {},
	"ELSE":              {},
	"EMPTY":             {},
	"ENABLE":            {},
	"ENCODING":          {},
	"ENCRYPTED":         {},
	"END":               {},
	"ENFORCED":          {},
	"ENUM":              {},
	"ERROR":             {},
	"ESCAPE":            {},
	"EVENT":             {},
	"EXCEPT":            {},
	"EXCLUDE":           {},
	"EXCLUDING":         {},
	"EXCLUSIVE":         {},
	"EXECUTE":           {},
	"EXISTS":            {},
	"EXPLAIN":           {},
	"EXPRESSION":        {},
	"EXTENSION":         {},
	"EXTERNAL":          {},
	"EXTRACT":           {},
	"FALSE":             {},
	"FAMILY":            {},
	"FETCH":             {},
	"FILTER":            {},
	"FINALIZE":          {},
	"FIRST":             {},
	"FLOAT":             {},
	"FOLLOWING":         {},
	"FOR":               {},
	"FORCE":             {},
	"FOREIGN":           {},
	"FORMAT":            {},
	"FORWARD":           {},
	"FREEZE":            {},
	"FROM":              {},
	"FULL":              {},
	"FUNCTION":          {},
	"FUNCTIONS":         {},
	"GENERATED":         {},
	"GLOBAL":            {},
	"GRANT":             {},
	"GRANTED":           {},
	"GREATEST":          {},
	"GROUP":             {},
	"GROUPING":          {},
	"GROUPS":            {},
	"HANDLER":           {},
	"HAVING":            {},
	"HEADER":            {},
	"HOLD":              {},
	"HOUR":              {},
	"IDENTITY":          {},
	"IF":                {},
	"ILIKE":             {},
	"IMMEDIATE":         {},
	"IMMUTABLE":         {},
	"IMPLICIT":          {},
	"IMPORT":            {},
	"IN":                {},
	"INCLUDE":           {},
	"INCLUDING":         {},
	"INCREMENT":         {},
	"INDENT":            {},
	"INDEX":             {},
	"INDEXES":           {},
	"INHERIT":           {},
	"INHERITS":          {},
	"INITIALLY":         {},
	"INLINE":            {},
	"INNER":             {},
	"INOUT":             {},
	"INPUT":             {},
	"INSENSITIVE":       {},
	"INSERT":            {},
	"INSTEAD":           {},
	"INT":               {},
	"INTEGER":           {},
	"INTERSECT":         {},
	"INTERVAL":          {},
	"INTO":              {},
	"INVOKER":           {},
	"IS":                {},
	"ISNULL":            {},
	"ISOLATION":         {},
	"JOIN":              {},
	"JSON":              {},
	"JSON_ARRAY":        {},
	"JSON_ARRAYAGG":     {},
	"JSON_EXISTS":       {},
	"JSON_OBJECT":       {},
	"JSON_OBJECTAGG":    {},
	"JSON_QUERY":        {},
	"JSON_SCALAR":       {},
	"JSON_SERIALIZE":    {},
	"JSON_TABLE":        {},
	"JSON_VALUE":        {},
	"KEEP":              {},
	"KEY":               {},
	"KEYS":              {},
	"LABEL":             {},
	"LANGUAGE":          {},
	"LARGE":             {},
	"LAST":              {},
	"LATERAL":           {},
	"LEADING":           {},
	"LEAKPROOF":         {},
	"LEAST":             {},
	"LEFT":              {},
	"LEVEL":             {},
	"LIKE":              {},
	"LIMIT":             {},
	"LISTEN":            {},
	"LOAD":              {},
	"LOCAL":             {},
	"LOCALTIME":         {},
	"LOCALTIMESTAMP":    {},
	"LOCATION":          {},
	"LOCK":              {},
	"LOCKED":            {},
	"LOGGED":            {},
	"MAPPING":           {},
	"MATCH":             {},
	"MATCHED":           {},
	"MATERIALIZED":      {},
	"MAXVALUE":          {},
	"MERGE":             {},
	"MERGE_ACTION":      {},
	"METHOD":            {},
	"MINUTE":            {},
	"MINVALUE":          {},
	"MODE":              {},
	"MONTH":             {},
	"MOVE":              {},
	"NAME":              {},
	"NAMES":             {},
	"NATIONAL":          {},
	"NATURAL":           {},
	"NCHAR":             {},
	"NESTED":            {},
	"NEW":               {},
	"NEXT":              {},
	"NFC":               {},
	"NFD":               {},
	"NFKC":              {},
	"NFKD":              {},
	"NO":                {},
	"NONE":              {},
	"NORMALIZE":         {},
	"NORMALIZED":        {},
	"NOT":               {},
	"NOTHING":           {},
	"NOTIFY":            {},
	"NOTNULL":           {},
	"NOWAIT":            {},
	"NULL":              {},
	"NULLIF":            {},
	"NULLS":             {},
	"NUMERIC":           {},
	"OBJECT":            {},
	"OBJECTS":           {},
	"OF":                {},
	"OFF":               {},
	"OFFSET":            {},
	"OIDS":              {},
	"OLD":               {},
	"OMIT":              {},
	"ON":                {},
	"ONLY":              {},
	"OPERATOR":          {},
	"OPTION":            {},
	"OPTIONS":           {},
	"OR":                {},
	"ORDER":             {},
	"ORDINALITY":        {},
	"OTHERS":            {},
	"OUT":               {},
	"OUTER":             {},
	"OVER":              {},
	"OVERLAPS":          {},
	"OVERLAY":           {},
	"OVERRIDING":        {},
	"OWNED":             {},
	"OWNER":             {},
	"PARALLEL":          {},
	"PARAMETER":         {},
	"PARSER":            {},
	"PARTIAL":           {},
	"PARTITION":         {},
	"PASSING":           {},
	"PASSWORD":          {},
	"PATH":              {},
	"PERIOD":            {},
	"PLACING":           {},
	"PLAN":              {},
	"PLANS":             {},
	"POLICY":            {},
	"POSITION":          {},
	"PRECEDING":         {},
	"PRECISION":         {},
	"PREPARE":           {},
	"PREPARED":          {},
	"PRESERVE":          {},
	"PRIMARY":           {},
	"PRIOR":             {},
	"PRIVILEGES":        {},
	"PROCEDURAL":        {},
	"PROCEDURE":         {},
	"PROCEDURES":        {},
	"PROGRAM":           {},
	"PUBLICATION":       {},
	"QUOTE":             {},
	"QUOTES":            {},
	"RANGE":             {},
	"READ":              {},
	"REAL":              {},
	"REASSIGN":          {},
	"RECURSIVE":         {},
	"REF":               {},
	"REFERENCES":        {},
	"REFERENCING":       {},
	"REFRESH":           {},
	"REINDEX":           {},
	"RELATIVE":          {},
	"RELEASE":           {},
	"RENAME":            {},
	"REPEATABLE":        {},
	"REPLACE":           {},
	"REPLICA":           {},
	"RESET":             {},
	"RESTART":           {},
	"RESTRICT":          {},
	"RETURN":            {},
	"RETURNING":         {},
	"RETURNS":           {},
	"REVOKE":            {},
	"RIGHT":             {},
	"ROLE":              {},
	"ROLLBACK":          {},
	"ROLLUP":            {},
	"ROUTINE":           {},
	"ROUTINES":          {},
	"ROW":               {},
	"ROWS":              {},
	"RULE":              {},
	"SAVEPOINT":         {},
	"SCALAR":            {},
	"SCHEMA":            {},
	"SCHEMAS":           {},
	"SCROLL":            {},
	"SEARCH":            {},
	"SECOND":            {},
	"SECURITY":          {},
	"SELECT":            {},
	"SEQUENCE":          {},
	"SEQUENCES":         {},
	"SERIALIZABLE":      {},
	"SERVER":            {},
	"SESSION":           {},
	"SESSION_USER":      {},
	"SET":               {},
	"SETOF":             {},
	"SETS":              {},
	"SHARE":             {},
	"SHOW":              {},
	"SIMILAR":           {},
	"SIMPLE":            {},
	"SKIP":              {},
	"SMALLINT":          {},
	"SNAPSHOT":          {},
	"SOME":              {},
	"SOURCE":            {},
	"SQL":               {},
	"STABLE":            {},
	"STANDALONE":        {},
	"START":             {},
	"STATEMENT":         {},
	"STATISTICS":        {},
	"STDIN":             {},
	"STDOUT":            {},
	"STORAGE":           {},
	"STORED":            {},
	"STRICT":            {},
	"STRING":            {},
	"STRIP":             {},
	"SUBSCRIPTION":      {},
	"SUBSTRING":         {},
	"SUPPORT":           {},
	"SYMMETRIC":         {},
	"SYSID":             {},
	"SYSTEM":            {},
	"SYSTEM_USER":       {},
	"TABLE":             {},
	"TABLES":            {},
	"TABLESAMPLE":       {},
	"TABLESPACE":        {},
	"TARGET":            {},
	"TEMP":              {},
	"TEMPLATE":          {},
	"TEMPORARY":         {},
	"TEXT":              {},
	"THEN":              {},
	"TIES":              {},
	"TIME":              {},
	"TIMESTAMP":         {},
	"TO":                {},
	"TRAILING":          {},
	"TRANSACTION":       {},
	"TRANSFORM":         {},
	"TREAT":             {},
	"TRIGGER":           {},
	"TRIM":              {},
	"TRUE":              {},
	"TRUNCATE":          {},
	"TRUSTED":           {},
	"TYPE":              {},
	"TYPES":             {},
	"UESCAPE":           {},
	"UNBOUNDED":         {},
	"UNCOMMITTED":       {},
	"UNCONDITIONAL":     {},
	"UNENCRYPTED":       {},
	"UNION":             {},
	"UNIQUE":            {},
	"UNKNOWN":           {},
	"UNLISTEN":          {},
	"UNLOGGED":          {},
	"UNTIL":             {},
	"UPDATE":            {},
	"USER":              {},
	"USING":             {},
	"VACUUM":            {},
	"VALID":             {},
	"VALIDATE":          {},
	"VALIDATOR":         {},
	"VALUE":             {},
	"VALUES":            {},
	"VARCHAR":           {},
	"VARIADIC":          {},
	"VARYING":           {},
	"VERBOSE":           {},
	"VERSION":           {},
	"VIEW":              {},
	"VIEWS":             {},
	"VIRTUAL":           {},
	"VOLATILE":          {},
	"WHEN":              {},
	"WHERE":             {},
	"WHITESPACE":        {},
	"WINDOW":            {},
	"WITH":              {},
	"WITHIN":            {},
	"WITHOUT":           {},
	"WORK":              {},
	"WRAPPER":           {},
	"WRITE":             {},
	"XML":               {},
	"XMLATTRIBUTES":     {},
	"XMLCONCAT":         {},
	"XMLELEMENT":        {},
	"XMLEXISTS":         {},
	"XMLFOREST":         {},
	"XMLNAMESPACES":     {},
	"XMLPARSE":          {},
	"XMLPI":             {},
	"XMLROOT":           {},
	"XMLSERIALIZE":      {},
	"XMLTABLE":          {},
	"YEAR":              {},
	"YES":               {},
	"ZONE":              {},
}
