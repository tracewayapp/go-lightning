package lit

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type Driver int

const (
	PostgreSQL Driver = iota
	MySQL
	SQLite
)

func (d Driver) String() string {
	switch d {
	case PostgreSQL:
		return "PostgreSQL"
	case MySQL:
		return "MySQL"
	case SQLite:
		return "SQLite"
	default:
		return "Unknown"
	}
}

func (d Driver) InsertAndGetId(ex Executor, query string, args ...any) (int, error) {
	switch d {
	case PostgreSQL:
		row := ex.QueryRow(query, args...)
		var id int
		err := row.Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	case MySQL:
		result, err := ex.Exec(query, args...)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(id), nil
	case SQLite:
		result, err := ex.Exec(query, args...)
		if err != nil {
			return 0, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(id), nil
	default:
		return 0, fmt.Errorf("unsupported driver: %v", d)
	}
}

type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type DbNamingStrategy interface {
	GetTableNameFromStructName(string) string
	GetColumnNameFromStructName(string) string
}

type DefaultDbNamingStrategy struct{}

func (d DefaultDbNamingStrategy) GetTableNameFromStructName(input string) string {
	return toSnakeCase(input) + "s"
}

func (d DefaultDbNamingStrategy) GetColumnNameFromStructName(input string) string {
	return toSnakeCase(input)
}

// toSnakeCase converts a CamelCase string to snake_case, keeping consecutive
// uppercase letters together as acronyms (e.g., "HTTPRequest" -> "http_request").
func toSnakeCase(input string) string {
	var result strings.Builder
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsUpper(r) {
			if i > 0 {
				prevLower := unicode.IsLower(runes[i-1])
				nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				prevUpper := unicode.IsUpper(runes[i-1])

				// Add underscore if:
				// - Previous char was lowercase (start of new word), OR
				// - Previous char was uppercase AND next char is lowercase (end of acronym)
				if prevLower || (prevUpper && nextLower) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

type FieldMap struct {
	ColumnsMap    map[string]int
	ColumnKeys    []string
	HasIntId      bool
	InsertQuery   string
	UpdateQuery   string
	InsertColumns []string
	Driver        Driver
}

type InsertUpdateQueryGenerator interface {
	GenerateInsertQuery(tableName string, columnKeys []string, hasIntId bool) (string, []string)
	GenerateUpdateQuery(tableName string, columnKeys []string) string
}

var StructToFieldMap = make(map[reflect.Type]*FieldMap)
var defaultDriver *Driver = nil

func RegisterDriver(driver Driver) {
	defaultDriver = &driver
}

func RegisterModel[T any](driver ...Driver) {
	var d Driver
	if len(driver) > 0 {
		d = driver[0]
	} else if defaultDriver != nil {
		d = *defaultDriver
	} else {
		panic("no driver provided and no default driver set.")
	}
	RegisterModelWithNaming[T](d, DefaultDbNamingStrategy{})
}

func RegisterModelWithNaming[T any](driver Driver, namingStrategy DbNamingStrategy) {
	t := reflect.TypeFor[T]()

	columnsMap := make(map[string]int)
	columnKeys := []string{}
	hasIntId := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Tag.Get("lit")
		if name == "" {
			name = namingStrategy.GetColumnNameFromStructName(field.Name)
		}
		if name == "id" {
			if field.Type.AssignableTo(reflect.TypeOf(0)) {
				hasIntId = true
			}
		}
		columnKeys = append(columnKeys, name)
		columnsMap[name] = i
	}

	tableName := namingStrategy.GetTableNameFromStructName(t.Name())

	var queryGenerator InsertUpdateQueryGenerator
	switch driver {
	case PostgreSQL:
		queryGenerator = PgInsertUpdateQueryGenerator{}
	case MySQL:
		queryGenerator = MySqlInsertUpdateQueryGenerator{}
	case SQLite:
		queryGenerator = SqliteInsertUpdateQueryGenerator{}
	default:
		panic(fmt.Sprintf("unsupported driver: %v", driver))
	}

	insertQuery, insertColumns := queryGenerator.GenerateInsertQuery(tableName, columnKeys, hasIntId)
	updateQuery := queryGenerator.GenerateUpdateQuery(tableName, columnKeys)

	StructToFieldMap[t] = &FieldMap{
		ColumnsMap:    columnsMap,
		ColumnKeys:    columnKeys,
		HasIntId:      hasIntId,
		InsertQuery:   insertQuery,
		UpdateQuery:   updateQuery,
		InsertColumns: insertColumns,
		Driver:        driver,
	}
}

func GetFieldMap(t reflect.Type) (*FieldMap, error) {
	val, ok := StructToFieldMap[t]
	if !ok {
		return nil, fmt.Errorf("non registered model %s used. Please call `lit.RegisterModel[%s](driver)` after you define %s", t.Name(), t.Name(), t.Name())
	}
	return val, nil
}
