package lit

import (
	"errors"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
)

func ValidateColumns[T any](columns []string, fieldMap *FieldMap) error {
	for _, column := range columns {
		if !slices.Contains(fieldMap.ColumnKeys, column) {
			return errors.New("invalid column that is not found in the struct: " + column)
		}
	}
	return nil
}

func GetPointersForColumns[T any](columns []string, fieldMap *FieldMap, t *T) *[]interface{} {
	var dest []interface{}

	for _, column := range columns {
		pos := fieldMap.ColumnsMap[column]
		dest = append(dest, reflect.ValueOf(t).Elem().Field(pos).Addr().Interface())
	}
	return &dest
}

func Select[T any](ex Executor, query string, args ...any) ([]*T, error) {
	rows, err := ex.Query(query, args...)
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

func SelectSingle[T any](ex Executor, query string, args ...any) (*T, error) {
	l, err := Select[T](ex, query, args...)
	if err != nil {
		return nil, err
	}
	if len(l) > 0 {
		return l[0], nil
	}
	return nil, nil
}

func Insert[T any](ex Executor, t *T) (int, error) {
	tType := reflect.TypeOf(*t)
	fieldMap, err := GetFieldMap(tType)
	if err != nil {
		return 0, err
	}

	if err := ValidateColumns[T](fieldMap.InsertColumns, fieldMap); err != nil {
		return 0, err
	}

	pointers := *GetPointersForColumns(fieldMap.InsertColumns, fieldMap, t)

	return fieldMap.Driver.InsertAndGetId(ex, fieldMap.InsertQuery, pointers...)
}

func InsertGenericUuid[T any](ex Executor, t *T) (string, error) {
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

	_, err = ex.Exec(fieldMap.InsertQuery, *GetPointersForColumns[T](fieldMap.InsertColumns, fieldMap, t)...)
	if err != nil {
		return "", err
	}

	return newUuidString, nil
}

func InsertGenericExistingUuid[T any](ex Executor, t *T) error {
	tType := reflect.TypeOf(*t)
	fieldMap, err := GetFieldMap(tType)
	if err != nil {
		return err
	}

	if err := ValidateColumns[T](fieldMap.InsertColumns, fieldMap); err != nil {
		return err
	}

	_, err = ex.Exec(fieldMap.InsertQuery, *GetPointersForColumns[T](fieldMap.InsertColumns, fieldMap, t)...)
	return err
}

func Update[T any](ex Executor, t *T, where string, args ...any) error {
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

	finalWhere := where
	if fieldMap.Driver == PostgreSQL && strings.Contains(where, "$") {
		offset := strings.Count(fieldMap.UpdateQuery, "$")
		finalWhere = pgRenumberPlaceholders(where, offset)
	}

	_, err = ex.Exec(fieldMap.UpdateQuery+finalWhere, params...)
	return err
}

func Delete(ex Executor, query string, args ...any) error {
	_, err := ex.Exec(query, args...)
	return err
}

func SelectMultipleNative[T any](ex Executor, mapLine func(*interface{ Scan(...any) error }, *T) error, query string, args ...any) ([]*T, error) {
	rows, err := ex.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*T{}

	for rows.Next() {
		var t T
		var scanner interface{ Scan(...any) error } = rows
		if err := mapLine(&scanner, &t); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func InsertNative(ex Executor, query string, args ...any) (int, error) {
	result, err := ex.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func UpdateNative(ex Executor, query string, args ...any) error {
	_, err := ex.Exec(query, args...)
	return err
}
