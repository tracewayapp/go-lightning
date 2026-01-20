package lit

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestUser struct {
	Id        int
	FirstName string
	LastName  string
	Email     string
}

type TestProduct struct {
	Id    string
	Name  string
	Price int
}

func TestDriverString(t *testing.T) {
	assert.Equal(t, "PostgreSQL", PostgreSQL.String())
	assert.Equal(t, "MySQL", MySQL.String())
	assert.Equal(t, "Unknown", Driver(99).String())
}

func TestDefaultDbNamingStrategy_GetTableNameFromStructName(t *testing.T) {
	ns := DefaultDbNamingStrategy{}
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "users"},
		{"UserProfile", "user_profiles"},
		{"HTTPRequest", "h_t_t_p_requests"},
		{"A", "as"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ns.GetTableNameFromStructName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultDbNamingStrategy_GetColumnNameFromStructName(t *testing.T) {
	ns := DefaultDbNamingStrategy{}
	tests := []struct {
		input    string
		expected string
	}{
		{"Id", "id"},
		{"FirstName", "first_name"},
		{"email", "email"},
		{"HTTPCode", "h_t_t_p_code"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ns.GetColumnNameFromStructName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegisterModel_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())

	RegisterModel[TestUser](PostgreSQL)

	fieldMap, err := GetFieldMap(reflect.TypeFor[TestUser]())
	require.NoError(t, err)
	require.NotNil(t, fieldMap)

	assert.True(t, fieldMap.HasIntId)
	assert.Equal(t, PostgreSQL, fieldMap.Driver)
	assert.Contains(t, fieldMap.ColumnKeys, "id")
	assert.Contains(t, fieldMap.ColumnKeys, "first_name")
	assert.Contains(t, fieldMap.ColumnKeys, "last_name")
	assert.Contains(t, fieldMap.ColumnKeys, "email")

	assert.Contains(t, fieldMap.InsertQuery, "RETURNING id")
	assert.Contains(t, fieldMap.InsertQuery, "$1")
}

func TestRegisterModel_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())

	RegisterModel[TestUser](MySQL)

	fieldMap, err := GetFieldMap(reflect.TypeFor[TestUser]())
	require.NoError(t, err)
	require.NotNil(t, fieldMap)

	assert.True(t, fieldMap.HasIntId)
	assert.Equal(t, MySQL, fieldMap.Driver)

	assert.NotContains(t, fieldMap.InsertQuery, "RETURNING")
	assert.Contains(t, fieldMap.InsertQuery, "?")
	assert.NotContains(t, fieldMap.InsertQuery, "$")
}

func TestGetFieldMap_NotRegistered(t *testing.T) {
	type UnregisteredType struct {
		Id int
	}
	delete(StructToFieldMap, reflect.TypeFor[UnregisteredType]())

	fieldMap, err := GetFieldMap(reflect.TypeFor[UnregisteredType]())
	assert.Error(t, err)
	assert.Nil(t, fieldMap)
	assert.Contains(t, err.Error(), "non registered model")
}

func TestPgInsertUpdateQueryGenerator_GenerateInsertQuery(t *testing.T) {
	gen := PgInsertUpdateQueryGenerator{}

	tests := []struct {
		name             string
		tableName        string
		columnKeys       []string
		hasIntId         bool
		expectedContains []string
		expectedColumns  []string
	}{
		{
			name:             "with int id",
			tableName:        "users",
			columnKeys:       []string{"id", "first_name", "last_name"},
			hasIntId:         true,
			expectedContains: []string{"INSERT INTO", "users", "DEFAULT", "RETURNING id", "$1", "$2"},
			expectedColumns:  []string{"first_name", "last_name"},
		},
		{
			name:             "without int id",
			tableName:        "products",
			columnKeys:       []string{"product_id", "name", "price"},
			hasIntId:         false,
			expectedContains: []string{"INSERT INTO", "products", "$1", "$2", "$3", "RETURNING id"},
			expectedColumns:  []string{"product_id", "name", "price"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, columns := gen.GenerateInsertQuery(tt.tableName, tt.columnKeys, tt.hasIntId)

			for _, s := range tt.expectedContains {
				assert.Contains(t, query, s)
			}
			assert.Equal(t, tt.expectedColumns, columns)
		})
	}
}

func TestPgInsertUpdateQueryGenerator_GenerateUpdateQuery(t *testing.T) {
	gen := PgInsertUpdateQueryGenerator{}

	columnKeys := []string{"id", "first_name", "last_name"}
	query := gen.GenerateUpdateQuery("users", columnKeys)

	assert.Contains(t, query, "UPDATE users")
	assert.Contains(t, query, "SET")
	assert.Contains(t, query, "id = $1")
	assert.Contains(t, query, "first_name = $2")
	assert.Contains(t, query, "last_name = $3")
	assert.Contains(t, query, "WHERE")
}

func TestMySqlInsertUpdateQueryGenerator_GenerateInsertQuery(t *testing.T) {
	gen := MySqlInsertUpdateQueryGenerator{}

	tests := []struct {
		name             string
		tableName        string
		columnKeys       []string
		hasIntId         bool
		expectedContains []string
		expectedColumns  []string
	}{
		{
			name:             "with int id",
			tableName:        "users",
			columnKeys:       []string{"id", "first_name", "last_name"},
			hasIntId:         true,
			expectedContains: []string{"INSERT INTO", "users", "NULL", "?"},
			expectedColumns:  []string{"first_name", "last_name"},
		},
		{
			name:             "without int id",
			tableName:        "products",
			columnKeys:       []string{"product_id", "name", "price"},
			hasIntId:         false,
			expectedContains: []string{"INSERT INTO", "products", "?"},
			expectedColumns:  []string{"product_id", "name", "price"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, columns := gen.GenerateInsertQuery(tt.tableName, tt.columnKeys, tt.hasIntId)

			for _, s := range tt.expectedContains {
				assert.Contains(t, query, s)
			}
			assert.NotContains(t, query, "RETURNING")
			assert.Equal(t, tt.expectedColumns, columns)
		})
	}
}

func TestMySqlInsertUpdateQueryGenerator_GenerateUpdateQuery(t *testing.T) {
	gen := MySqlInsertUpdateQueryGenerator{}

	columnKeys := []string{"id", "first_name", "last_name"}
	query := gen.GenerateUpdateQuery("users", columnKeys)

	assert.Contains(t, query, "UPDATE users")
	assert.Contains(t, query, "SET")
	assert.Contains(t, query, "id = ?")
	assert.Contains(t, query, "first_name = ?")
	assert.Contains(t, query, "last_name = ?")
	assert.Contains(t, query, "WHERE")
	assert.NotContains(t, query, "$")
}

func TestJoinForIn(t *testing.T) {
	tests := []struct {
		name     string
		ids      []int
		expected string
	}{
		{"empty", []int{}, ""},
		{"single", []int{1}, "1"},
		{"multiple", []int{1, 2, 3}, "1,2,3"},
		{"negative", []int{-1, 0, 1}, "-1,0,1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinForIn(tt.ids)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinStringForIn_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	tests := []struct {
		name     string
		offset   int
		params   []string
		expected string
	}{
		{"empty", 0, []string{}, ""},
		{"no offset", 0, []string{"a", "b"}, "$1,$2"},
		{"with offset", 2, []string{"a", "b"}, "$3,$4"},
		{"large offset", 10, []string{"x"}, "$11"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinStringForIn[TestUser](tt.offset, tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinStringForIn_MySQL(t *testing.T) {
	type MySQLUser struct {
		Id   int
		Name string
	}
	delete(StructToFieldMap, reflect.TypeFor[MySQLUser]())
	RegisterModel[MySQLUser](MySQL)

	tests := []struct {
		name     string
		offset   int
		params   []string
		expected string
	}{
		{"empty", 0, []string{}, ""},
		{"ignores offset", 5, []string{"a", "b"}, "?,?"},
		{"multiple", 0, []string{"x", "y", "z"}, "?,?,?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinStringForIn[MySQLUser](tt.offset, tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinStringForInWithDriver(t *testing.T) {
	assert.Equal(t, "$1,$2", JoinStringForInWithDriver(PostgreSQL, 0, 2))
	assert.Equal(t, "$3,$4,$5", JoinStringForInWithDriver(PostgreSQL, 2, 3))

	assert.Equal(t, "?,?", JoinStringForInWithDriver(MySQL, 0, 2))
	assert.Equal(t, "?,?,?", JoinStringForInWithDriver(MySQL, 999, 3))
}

func TestPgRenumberPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		where    string
		offset   int
		expected string
	}{
		{"no placeholders", "id = 5", 3, "id = 5"},
		{"single placeholder", "id = $1", 3, "id = $4"},
		{"multiple placeholders", "id = $1 AND status = $2", 5, "id = $6 AND status = $7"},
		{"placeholder at end", "name = $1", 2, "name = $3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgRenumberPlaceholders(tt.where, tt.offset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSelect_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
		AddRow(1, "John", "Doe", "john@example.com").
		AddRow(2, "Jane", "Smith", "jane@example.com")

	mock.ExpectQuery("SELECT \\* FROM test_users").WillReturnRows(rows)

	users, err := Select[TestUser](db, "SELECT * FROM test_users")
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "John", users[0].FirstName)
	assert.Equal(t, "Jane", users[1].FirstName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSelect_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
		AddRow(1, "John", "Doe", "john@example.com").
		AddRow(2, "Jane", "Smith", "jane@example.com")

	mock.ExpectQuery("SELECT \\* FROM test_users").WillReturnRows(rows)

	users, err := Select[TestUser](db, "SELECT * FROM test_users")
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "John", users[0].FirstName)
	assert.Equal(t, "Jane", users[1].FirstName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSelectSingle_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
		AddRow(1, "John", "Doe", "john@example.com")

	mock.ExpectQuery("SELECT \\* FROM test_users WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	user, err := SelectSingle[TestUser](db, "SELECT * FROM test_users WHERE id = $1", 1)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "John", user.FirstName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSelectSingle_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
		AddRow(1, "John", "Doe", "john@example.com")

	mock.ExpectQuery("SELECT \\* FROM test_users WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(rows)

	user, err := SelectSingle[TestUser](db, "SELECT * FROM test_users WHERE id = ?", 1)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "John", user.FirstName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSelectSingle_NoResults_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"})

	mock.ExpectQuery("SELECT \\* FROM test_users WHERE id = \\$1").
		WithArgs(999).
		WillReturnRows(rows)

	user, err := SelectSingle[TestUser](db, "SELECT * FROM test_users WHERE id = $1", 999)
	require.NoError(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSelectSingle_NoResults_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"})

	mock.ExpectQuery("SELECT \\* FROM test_users WHERE id = \\?").
		WithArgs(999).
		WillReturnRows(rows)

	user, err := SelectSingle[TestUser](db, "SELECT * FROM test_users WHERE id = ?", 999)
	require.NoError(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsert_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(42)

	mock.ExpectQuery("INSERT INTO test_users").
		WithArgs("John", "Doe", "john@example.com").
		WillReturnRows(rows)

	user := &TestUser{FirstName: "John", LastName: "Doe", Email: "john@example.com"}
	id, err := Insert[TestUser](db, user)
	require.NoError(t, err)
	assert.Equal(t, 42, id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsert_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO test_users").
		WithArgs("John", "Doe", "john@example.com").
		WillReturnResult(sqlmock.NewResult(42, 1))

	user := &TestUser{FirstName: "John", LastName: "Doe", Email: "john@example.com"}
	id, err := Insert[TestUser](db, user)
	require.NoError(t, err)
	assert.Equal(t, 42, id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("UPDATE test_users SET").
		WithArgs(1, "John", "Doe", "john@example.com", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	user := &TestUser{Id: 1, FirstName: "John", LastName: "Doe", Email: "john@example.com"}
	err = Update[TestUser](db, user, "id = $1", 1)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("UPDATE test_users SET").
		WithArgs(1, "John", "Doe", "john@example.com", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	user := &TestUser{Id: 1, FirstName: "John", LastName: "Doe", Email: "john@example.com"}
	err = Update[TestUser](db, user, "id = ?", 1)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_NoWhere(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &TestUser{Id: 1, FirstName: "John", LastName: "Doe", Email: "john@example.com"}
	err = Update[TestUser](db, user, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "where")
}

func TestDelete_PostgreSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM test_users WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Delete(db, "DELETE FROM test_users WHERE id = $1", 1)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_MySQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM test_users WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Delete(db, "DELETE FROM test_users WHERE id = ?", 1)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutorWithTransaction_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
		AddRow(1, "John", "Doe", "john@example.com")

	mock.ExpectQuery("SELECT \\* FROM test_users").WillReturnRows(rows)
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	users, err := Select[TestUser](tx, "SELECT * FROM test_users")
	require.NoError(t, err)
	assert.Len(t, users, 1)

	err = tx.Commit()
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecutorWithTransaction_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())
	RegisterModel[TestUser](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
		AddRow(1, "John", "Doe", "john@example.com")

	mock.ExpectQuery("SELECT \\* FROM test_users").WillReturnRows(rows)
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	users, err := Select[TestUser](tx, "SELECT * FROM test_users")
	require.NoError(t, err)
	assert.Len(t, users, 1)

	err = tx.Commit()
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertGenericUuid_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestProduct]())
	RegisterModel[TestProduct](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO test_products").
		WillReturnResult(sqlmock.NewResult(0, 1))

	product := &TestProduct{Name: "Widget", Price: 100}
	uuid, err := InsertGenericUuid[TestProduct](db, product)
	require.NoError(t, err)
	assert.NotEmpty(t, uuid)
	assert.Equal(t, uuid, product.Id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertGenericUuid_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestProduct]())
	RegisterModel[TestProduct](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO test_products").
		WillReturnResult(sqlmock.NewResult(0, 1))

	product := &TestProduct{Name: "Widget", Price: 100}
	uuid, err := InsertGenericUuid[TestProduct](db, product)
	require.NoError(t, err)
	assert.NotEmpty(t, uuid)
	assert.Equal(t, uuid, product.Id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertGenericExistingUuid_PostgreSQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestProduct]())
	RegisterModel[TestProduct](PostgreSQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO test_products").
		WithArgs("existing-uuid-123", "Widget", 100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	product := &TestProduct{Id: "existing-uuid-123", Name: "Widget", Price: 100}
	err = InsertGenericExistingUuid[TestProduct](db, product)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertGenericExistingUuid_MySQL(t *testing.T) {
	delete(StructToFieldMap, reflect.TypeFor[TestProduct]())
	RegisterModel[TestProduct](MySQL)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO test_products").
		WithArgs("existing-uuid-123", "Widget", 100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	product := &TestProduct{Id: "existing-uuid-123", Name: "Widget", Price: 100}
	err = InsertGenericExistingUuid[TestProduct](db, product)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
