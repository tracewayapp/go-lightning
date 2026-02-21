package lit

import (
	"reflect"
	"strconv"
	"strings"
)

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

func JoinStringForIn[T any](offset int, params []string) string {
	fieldMap, err := GetFieldMap(reflect.TypeFor[T]())
	if err != nil {
		return pgJoinStringForIn(offset, len(params))
	}

	switch fieldMap.Driver {
	case PostgreSQL:
		return pgJoinStringForIn(offset, len(params))
	case MySQL:
		return mysqlJoinStringForIn(len(params))
	case SQLite:
		return sqliteJoinStringForIn(len(params))
	default:
		return pgJoinStringForIn(offset, len(params))
	}
}

func JoinStringForInWithDriver(driver Driver, offset int, count int) string {
	switch driver {
	case PostgreSQL:
		return pgJoinStringForIn(offset, count)
	case MySQL:
		return mysqlJoinStringForIn(count)
	case SQLite:
		return sqliteJoinStringForIn(count)
	default:
		return pgJoinStringForIn(offset, count)
	}
}
