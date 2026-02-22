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

	return fieldMap.Driver.JoinStringForIn(offset, len(params))
}

func JoinStringForInWithDriver(driver Driver, offset int, count int) string {
	return driver.JoinStringForIn(offset, count)
}
