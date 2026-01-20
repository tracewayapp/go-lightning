package lightning

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

type DbNamingStrategy interface {
	GetTableNameFromStructName(string) string
	GetColumnNameFromStructName(string) string
}

type DefaultDbNamingStrategy struct{}

func (d DefaultDbNamingStrategy) GetTableNameFromStructName(input string) string {
	var result strings.Builder
	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	result.WriteRune('s')
	return result.String()
}

func (d DefaultDbNamingStrategy) GetColumnNameFromStructName(input string) string {
	var result strings.Builder
	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func SelectMultipleNative[T any](tx *sql.Tx, mapLine func(*sql.Rows, *T) error, query string, args ...any) ([]*T, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*T{}

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var t T
		if err := mapLine(rows, &t); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}
func SelectSingleNative[T any](tx *sql.Tx, mapLine func(*sql.Rows, *T) error, query string, args ...any) (*T, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var t T
		if err := mapLine(rows, &t); err != nil {
			return nil, err
		}
		return &t, nil
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func InsertNative(tx *sql.Tx, query string, args ...any) (int, error) {
	result, err := tx.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func UpdateNative(tx *sql.Tx, query string, args ...any) error {
	_, err := (*tx).Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func Delete(tx *sql.Tx, query string, args ...any) error {
	_, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}
func ValidateColumns[T any](columns []string, fieldMap *FieldMap) error {
	for _, column := range columns {
		if !slices.Contains((*fieldMap).ColumnKeys, column) {
			return errors.New("invalid column that is not found in the struct " + column)
		}
	}
	return nil
}

func GetPointersForColumns[T any](columns []string, fieldMap *FieldMap, t *T) *[]interface{} {
	var dest []interface{}

	// this function assumes that all paths that lead to it have validated the columns
	for _, column := range columns {
		pos := (*fieldMap).ColumnsMap[column]

		dest = append(dest, reflect.ValueOf(t).Elem().Field(pos).Addr().Interface())
	}
	return &dest
}

func InsertUuid[T any](tx *sql.Tx, t *T) (string, error) {
	tType := reflect.TypeOf(*t)
	fieldMap, err := GetFieldMap(tType)

	if err != nil {
		return "", err
	}

	newUuid, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	newUuidString := newUuid.String()
	reflect.ValueOf(t).Elem().Field(fieldMap.ColumnsMap["id"]).SetString(newUuidString)

	if err := ValidateColumns[T](fieldMap.InsertColumns, fieldMap); err != nil {
		return "", err
	}
	_, err = tx.Exec(
		fieldMap.InsertQuery,
		*GetPointersForColumns[T](fieldMap.InsertColumns, fieldMap, t)...,
	)
	if err != nil {
		return "", err
	}

	return newUuidString, nil
}

func Update[T any](tx *sql.Tx, t *T, where string, args ...any) error {
	if len(where) == 0 {
		return errors.New("parameter 'where' was not present")
	}
	tType := reflect.TypeOf(*t)
	fieldMap, err := GetFieldMap(tType)
	if err != nil {
		return err
	}

	if err := ValidateColumns[T](fieldMap.ColumnKeys, fieldMap); err != nil {
		return err
	}
	params := append(*GetPointersForColumns[T](fieldMap.ColumnKeys, fieldMap, t), args...)

	_, err = tx.Exec(
		fieldMap.UpdateQuery+where,
		params...,
	)
	if err != nil {
		return err
	}

	return nil
}

func SelectSingle[T any](tx *sql.Tx, query string, args ...any) (*T, error) {

	l, err := Select[T](tx, query, args...)

	if err != nil {
		return nil, err
	}
	if len(l) > 0 {
		return l[0], nil
	}

	return nil, nil
}

func Select[T any](tx *sql.Tx, query string, args ...any) ([]*T, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*T{}

	fieldMap, err := GetFieldMap(reflect.TypeFor[T]())
	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	if err := ValidateColumns[T](columns, fieldMap); err != nil {
		return nil, err
	}

	for rows.Next() {
		var t T
		if err := rows.Scan(*GetPointersForColumns[T](columns, fieldMap, &t)...); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func Insert[T any](tx *sql.Tx, t *T) (int, error) {
	tType := reflect.TypeOf(*t)
	fieldMap, err := GetFieldMap(tType)
	if err != nil {
		return 0, err
	}

	if err := ValidateColumns[T](fieldMap.InsertColumns, fieldMap); err != nil {
		return 0, err
	}

	result, err := tx.Exec(
		fieldMap.InsertQuery,
		*GetPointersForColumns[T](fieldMap.InsertColumns, fieldMap, t)...,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func InsertExistingUuid[T any](tx *sql.Tx, t *T) error {
	tType := reflect.TypeOf(*t)
	fieldMap, err := GetFieldMap(tType)
	if err != nil {
		return err
	}

	if err := ValidateColumns[T](fieldMap.InsertColumns, fieldMap); err != nil {
		return err
	}

	_, err = tx.Exec(
		fieldMap.InsertQuery,
		*GetPointersForColumns[T](fieldMap.InsertColumns, fieldMap, t)...,
	)
	return err
}

func JoinForIn(ids []int) string {
	var sb strings.Builder
	for index, id := range ids {
		sb.WriteString(strconv.Itoa(id))
		if index < len(ids)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

var StructToFieldMap = make(map[reflect.Type]*FieldMap)

type FieldMap struct {
	ColumnsMap map[string]int
	ColumnKeys []string
	HasIntId   bool

	InsertQuery   string
	UpdateQuery   string
	InsertColumns []string
}

type InsertUpdateQueryGenerator interface {
	GenerateInsertQuery(tableName string, columnKeys []string, hasIntId bool) (string, []string)
	GenerateUpdateQuery(tableName string, columnKeys []string) string
}

func Register[T any](namingStrategy DbNamingStrategy, queryGenerator InsertUpdateQueryGenerator) {
	t := reflect.TypeFor[T]()

	columnsMap := make(map[string]int)
	columnKeys := []string{}
	hasIntId := false
	for i := 0; i < t.NumField(); i++ {
		name := namingStrategy.GetColumnNameFromStructName(t.Field(i).Name)
		if name == "id" {
			if t.Field(i).Type.AssignableTo(reflect.TypeOf(0)) {
				hasIntId = true
			}
		}
		columnKeys = append(columnKeys, name)
		columnsMap[name] = i
	}

	tableName := namingStrategy.GetTableNameFromStructName(t.Name())

	insertQuery, insertColumns := queryGenerator.GenerateInsertQuery(tableName, columnKeys, hasIntId)
	updateQuery := queryGenerator.GenerateUpdateQuery(tableName, columnKeys)

	StructToFieldMap[t] = &FieldMap{
		ColumnsMap:    columnsMap,
		ColumnKeys:    columnKeys,
		HasIntId:      hasIntId,
		InsertQuery:   insertQuery,
		UpdateQuery:   updateQuery,
		InsertColumns: insertColumns,
	}
}

func GetFieldMap(t reflect.Type) (*FieldMap, error) {
	val, ok := StructToFieldMap[t]
	if !ok {
		return nil, fmt.Errorf("Non registered model %s used. Please call `var _ = Register[%s]()` after you define %s", t.Name(), t.Name(), t.Name())
	}
	return val, nil
}
