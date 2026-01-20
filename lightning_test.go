package lightning

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test struct for Register and GetFieldMap tests
type TestUser struct {
	Id        int
	FirstName string
	LastName  string
	Email     string
}

type TestProduct struct {
	ProductId   string
	ProductName string
	Price       float64
}

type TestUuidEntity struct {
	Id          string
	Name        string
	Description string
}

// ========== DefaultDbNamingStrategy Tests ==========

func TestGetTableNameFromStructName(t *testing.T) {
	strategy := DefaultDbNamingStrategy{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "user", "users"},
		{"simple capitalized", "User", "users"},
		{"camel case", "UserProfile", "user_profiles"},
		{"multiple words", "UserOrderHistory", "user_order_historys"},
		{"single uppercase", "A", "as"},
		{"all uppercase", "ABC", "a_b_cs"},
		{"empty string", "", "s"},
		{"lowercase with numbers", "user123", "user123s"},
		{"mixed case", "firstName", "first_names"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.GetTableNameFromStructName(tt.input)
			if result != tt.expected {
				t.Errorf("GetTableNameFromStructName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetColumnNameFromStructName(t *testing.T) {
	strategy := DefaultDbNamingStrategy{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "user", "user"},
		{"simple capitalized", "User", "user"},
		{"camel case", "FirstName", "first_name"},
		{"multiple words", "UserOrderHistory", "user_order_history"},
		{"single uppercase", "A", "a"},
		{"all uppercase", "ABC", "a_b_c"},
		{"empty string", "", ""},
		{"lowercase with numbers", "user123", "user123"},
		{"mixed case", "firstName", "first_name"},
		{"id field", "Id", "id"},
		{"trailing uppercase", "UserID", "user_i_d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.GetColumnNameFromStructName(tt.input)
			if result != tt.expected {
				t.Errorf("GetColumnNameFromStructName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ========== JoinForIn Tests ==========

func TestJoinForIn(t *testing.T) {
	tests := []struct {
		name     string
		ids      []int
		expected string
	}{
		{"empty ids", []int{}, ""},
		{"single id", []int{1}, "1"},
		{"multiple ids", []int{1, 2, 3}, "1,2,3"},
		{"large ids", []int{100, 200, 300}, "100,200,300"},
		{"negative ids", []int{-1, 0, 1}, "-1,0,1"},
		{"single zero", []int{0}, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinForIn(tt.ids)
			if result != tt.expected {
				t.Errorf("JoinForIn(%v) = %q, want %q", tt.ids, result, tt.expected)
			}
		})
	}
}

// ========== Register and GetFieldMap Tests ==========

func TestRegister(t *testing.T) {
	// Clear any previous registrations for clean test
	delete(StructToFieldMap, reflect.TypeFor[TestUser]())

	Register[TestUser](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	fieldMap, err := GetFieldMap(reflect.TypeFor[TestUser]())
	if err != nil {
		t.Fatalf("GetFieldMap failed after Register: %v", err)
	}

	// Verify ColumnsMap
	expectedColumns := map[string]int{
		"id":         0,
		"first_name": 1,
		"last_name":  2,
		"email":      3,
	}

	for col, idx := range expectedColumns {
		if fieldMap.ColumnsMap[col] != idx {
			t.Errorf("ColumnsMap[%q] = %d, want %d", col, fieldMap.ColumnsMap[col], idx)
		}
	}

	// Verify ColumnKeys
	expectedKeys := []string{"id", "first_name", "last_name", "email"}
	if len(fieldMap.ColumnKeys) != len(expectedKeys) {
		t.Errorf("ColumnKeys length = %d, want %d", len(fieldMap.ColumnKeys), len(expectedKeys))
	}
	for i, key := range expectedKeys {
		if fieldMap.ColumnKeys[i] != key {
			t.Errorf("ColumnKeys[%d] = %q, want %q", i, fieldMap.ColumnKeys[i], key)
		}
	}

	// Verify HasIntId
	if !fieldMap.HasIntId {
		t.Error("HasIntId = false, want true")
	}

	// Verify InsertQuery contains expected parts
	if fieldMap.InsertQuery == "" {
		t.Error("InsertQuery is empty")
	}
	if !containsStr(fieldMap.InsertQuery, "INSERT INTO") {
		t.Error("InsertQuery missing 'INSERT INTO'")
	}
	if !containsStr(fieldMap.InsertQuery, "test_users") {
		t.Errorf("InsertQuery missing table name 'test_users': %s", fieldMap.InsertQuery)
	}
	if !containsStr(fieldMap.InsertQuery, "RETURNING id") {
		t.Error("InsertQuery missing 'RETURNING id'")
	}
	if !containsStr(fieldMap.InsertQuery, "DEFAULT") {
		t.Error("InsertQuery missing 'DEFAULT' for id column")
	}

	// Verify UpdateQuery contains expected parts
	if fieldMap.UpdateQuery == "" {
		t.Error("UpdateQuery is empty")
	}
	if !containsStr(fieldMap.UpdateQuery, "UPDATE") {
		t.Error("UpdateQuery missing 'UPDATE'")
	}
	if !containsStr(fieldMap.UpdateQuery, "test_users") {
		t.Errorf("UpdateQuery missing table name 'test_users': %s", fieldMap.UpdateQuery)
	}
	if !containsStr(fieldMap.UpdateQuery, "SET") {
		t.Error("UpdateQuery missing 'SET'")
	}
	if !containsStr(fieldMap.UpdateQuery, "WHERE") {
		t.Error("UpdateQuery missing 'WHERE'")
	}

	// Verify InsertColumns excludes 'id'
	for _, col := range fieldMap.InsertColumns {
		if col == "id" {
			t.Error("InsertColumns should not contain 'id' when HasIntId is true")
		}
	}
}

func TestRegisterWithoutIntId(t *testing.T) {
	// Clear any previous registrations for clean test
	delete(StructToFieldMap, reflect.TypeFor[TestProduct]())

	Register[TestProduct](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	fieldMap, err := GetFieldMap(reflect.TypeFor[TestProduct]())
	if err != nil {
		t.Fatalf("GetFieldMap failed after Register: %v", err)
	}

	// Verify HasIntId is false (ProductId is string, not int)
	if fieldMap.HasIntId {
		t.Error("HasIntId = true, want false (ProductId is string)")
	}

	// Verify InsertColumns includes all columns (no DEFAULT needed)
	expectedInsertColumns := []string{"product_id", "product_name", "price"}
	if len(fieldMap.InsertColumns) != len(expectedInsertColumns) {
		t.Errorf("InsertColumns length = %d, want %d", len(fieldMap.InsertColumns), len(expectedInsertColumns))
	}
}

func TestGetFieldMapUnregistered(t *testing.T) {
	// Define a type that's never registered
	type UnregisteredType struct {
		Field string
	}

	_, err := GetFieldMap(reflect.TypeFor[UnregisteredType]())
	if err == nil {
		t.Error("GetFieldMap should return error for unregistered type")
	}
}

// ========== Database Function Tests (require sql mock) ==========
// Note: SelectMultipleNative, SelectSingle, InsertNative, UpdateNative, Delete
// require a database connection or mock. These tests would need
// a testing database or sqlmock package for proper testing.

func TestSelectMultipleNative(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		expectedCount int
	}{
		{
			name:  "multiple rows returned",
			query: "SELECT id, first_name, last_name, email FROM users",
			args:  []any{},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					AddRow(1, "John", "Doe", "john@example.com").
					AddRow(2, "Jane", "Smith", "jane@example.com").
					AddRow(3, "Bob", "Johnson", "bob@example.com")
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			expectedError: false,
			expectedCount: 3,
		},
		{
			name:  "empty result set",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id > ?",
			args:  []any{1000},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"})
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(1000).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:  "single row",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id = ?",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					AddRow(1, "John", "Doe", "john@example.com")
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:  "query error",
			query: "SELECT id, first_name, last_name, email FROM users",
			args:  []any{},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:  "scan error",
			query: "SELECT id, first_name FROM users",
			args:  []any{},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Only 2 columns but mapTestUser expects 4
				rows := sqlmock.NewRows([]string{"id", "first_name"}).
					AddRow(1, "John")
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:  "rows error during iteration",
			query: "SELECT id, first_name, last_name, email FROM users",
			args:  []any{},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					AddRow(1, "John", "Doe", "john@example.com").
					RowError(0, sql.ErrConnDone)
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			result, err := SelectMultipleNative[TestUser](tx, mapTestUser, tt.query, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)

				// Additional validation for non-empty results
				if tt.expectedCount > 0 {
					assert.Equal(t, 1, result[0].Id)
					assert.Equal(t, "John", result[0].FirstName)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSelectSingleNative(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		expectNil     bool
		validateUser  func(*testing.T, *TestUser)
	}{
		{
			name:  "row found",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id = ?",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					AddRow(1, "John", "Doe", "john@example.com")
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectNil:     false,
			validateUser: func(t *testing.T, user *TestUser) {
				assert.Equal(t, 1, user.Id)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.LastName)
				assert.Equal(t, "john@example.com", user.Email)
			},
		},
		{
			name:  "no rows returns nil not error",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id = ?",
			args:  []any{999},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"})
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(999).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectNil:     true,
			validateUser:  nil,
		},
		{
			name:  "first row only when multiple exist",
			query: "SELECT id, first_name, last_name, email FROM users",
			args:  []any{},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					AddRow(1, "John", "Doe", "john@example.com").
					AddRow(2, "Jane", "Smith", "jane@example.com")
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			expectedError: false,
			expectNil:     false,
			validateUser: func(t *testing.T, user *TestUser) {
				assert.Equal(t, 1, user.Id)
				assert.Equal(t, "John", user.FirstName)
			},
		},
		{
			name:  "query error",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id = ?",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
			expectNil:     true,
			validateUser:  nil,
		},
		{
			name:  "scan error",
			query: "SELECT id, first_name FROM users WHERE id = ?",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Only 2 columns but mapTestUser expects 4
				rows := sqlmock.NewRows([]string{"id", "first_name"}).
					AddRow(1, "John")
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: true,
			expectNil:     true,
			validateUser:  nil,
		},
		{
			name:  "rows error after first row",
			query: "SELECT id, first_name, last_name, email FROM users",
			args:  []any{},
			setupMock: func(mock sqlmock.Sqlmock) {
				// CloseError simulates an error when closing rows (which triggers rows.Err())
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					CloseError(sql.ErrConnDone)
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			expectedError: true,
			expectNil:     true,
			validateUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			result, err := SelectSingleNative[TestUser](tx, mapTestUser, tt.query, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectNil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					if tt.validateUser != nil {
						tt.validateUser(t, result)
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestInsertNative(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedID    int
		expectedError bool
	}{
		{
			name:  "successful insert returns id",
			query: "INSERT INTO users (first_name, last_name, email) VALUES (?, ?, ?)",
			args:  []any{"John", "Doe", "john@example.com"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("John", "Doe", "john@example.com").
					WillReturnResult(sqlmock.NewResult(42, 1))
			},
			expectedID:    42,
			expectedError: false,
		},
		{
			name:  "large id value",
			query: "INSERT INTO users (first_name, last_name, email) VALUES (?, ?, ?)",
			args:  []any{"Jane", "Smith", "jane@example.com"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("Jane", "Smith", "jane@example.com").
					WillReturnResult(sqlmock.NewResult(2147483648, 1)) // > int32 max
			},
			expectedID:    2147483648,
			expectedError: false,
		},
		{
			name:  "parameterized query",
			query: "INSERT INTO users (first_name, last_name, email) VALUES (?, ?, ?)",
			args:  []any{"Bob", "Johnson", "bob@example.com"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("Bob", "Johnson", "bob@example.com").
					WillReturnResult(sqlmock.NewResult(100, 1))
			},
			expectedID:    100,
			expectedError: false,
		},
		{
			name:  "exec error",
			query: "INSERT INTO users (first_name, last_name, email) VALUES (?, ?, ?)",
			args:  []any{"John", "Doe", "john@example.com"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("John", "Doe", "john@example.com").
					WillReturnError(sql.ErrTxDone)
			},
			expectedID:    0,
			expectedError: true,
		},
		{
			name:  "last insert id error",
			query: "INSERT INTO users (first_name) VALUES (?)",
			args:  []any{"John"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("John").
					WillReturnResult(errorResult{err: fmt.Errorf("LastInsertId not supported")})
			},
			expectedID:    0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			id, err := InsertNative(tx, tt.query, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, 0, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateNative(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name:  "successful update",
			query: "UPDATE users SET first_name = ? WHERE id = ?",
			args:  []any{"Jane", 1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET first_name").
					WithArgs("Jane", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:  "update with where clause",
			query: "UPDATE users SET first_name = ?, last_name = ? WHERE id = ?",
			args:  []any{"John", "Smith", 5},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET (.+) WHERE id").
					WithArgs("John", "Smith", 5).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:  "zero rows affected is not error",
			query: "UPDATE users SET first_name = ? WHERE id = ?",
			args:  []any{"Jane", 999},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET first_name").
					WithArgs("Jane", 999).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
			expectedError: false,
		},
		{
			name:  "exec error",
			query: "UPDATE users SET first_name = ? WHERE id = ?",
			args:  []any{"Jane", 1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET first_name").
					WithArgs("Jane", 1).
					WillReturnError(sql.ErrTxDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			err := UpdateNative(tx, tt.query, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name:  "successful delete",
			query: "DELETE FROM users WHERE id = ?",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:  "delete with where clause",
			query: "DELETE FROM users WHERE email = ? AND id > ?",
			args:  []any{"test@example.com", 10},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE (.+)").
					WithArgs("test@example.com", 10).
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
			expectedError: false,
		},
		{
			name:  "zero rows affected is not error",
			query: "DELETE FROM users WHERE id = ?",
			args:  []any{999},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
			expectedError: false,
		},
		{
			name:  "exec error",
			query: "DELETE FROM users WHERE id = ?",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id").
					WithArgs(1).
					WillReturnError(sql.ErrTxDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			err := Delete(tx, tt.query, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Helper function
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ========== Database Test Helpers ==========

// setupMockDB creates a mock database and transaction for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *sql.Tx) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	return db, mock, tx
}

// mapTestUser is a mapper function for TestUser
func mapTestUser(rows *sql.Rows, user *TestUser) error {
	return rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email)
}

// mapTestProduct is a mapper function for TestProduct
func mapTestProduct(rows *sql.Rows, product *TestProduct) error {
	return rows.Scan(&product.ProductId, &product.ProductName, &product.Price)
}

// errorResult is a custom sql.Result for testing LastInsertId errors
type errorResult struct {
	err error
}

func (e errorResult) LastInsertId() (int64, error) {
	return 0, e.err
}

func (e errorResult) RowsAffected() (int64, error) {
	return 0, e.err
}

// mockQueryGenerator is a simple query generator for testing
type mockQueryGenerator struct{}

func (m mockQueryGenerator) GenerateInsertQuery(tableName string, columnKeys []string, hasIntId bool) (string, []string) {
	var insertQuery strings.Builder
	insertQuery.WriteString("INSERT INTO ")
	insertQuery.WriteString(tableName)
	insertQuery.WriteString(" (")

	totalKeys := len(columnKeys)
	for i, k := range columnKeys {
		insertQuery.WriteString(k)
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
			if i != totalKeys-1 {
				insertQuery.WriteString(",")
			}
		} else {
			insertColumns = append(insertColumns, k)
			insertQuery.WriteString("$" + strconv.Itoa(counter))
			if i != totalKeys-1 {
				insertQuery.WriteString(",")
			}
			counter++
		}
	}
	insertQuery.WriteString(") RETURNING id")

	return insertQuery.String(), insertColumns
}

func (m mockQueryGenerator) GenerateUpdateQuery(tableName string, columnKeys []string) string {
	var updateQuery strings.Builder
	updateQuery.WriteString("UPDATE ")
	updateQuery.WriteString(tableName)
	updateQuery.WriteString(" SET ")

	totalKeys := len(columnKeys)
	for i, k := range columnKeys {
		updateQuery.WriteString(k)
		updateQuery.WriteString(" = $" + strconv.Itoa(i+1))
		if i != totalKeys-1 {
			updateQuery.WriteString(",")
		}
	}

	updateQuery.WriteString(" WHERE ")

	return updateQuery.String()
}

func TestSelectSingle(t *testing.T) {
	Register[TestUser](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	tests := []struct {
		name          string
		query         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		expectNil     bool
	}{
		{
			name:  "row found",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id = $1",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					AddRow(1, "John", "Doe", "john@example.com")
				mock.ExpectQuery("SELECT (.+)").WithArgs(1).WillReturnRows(rows)
			},
			expectedError: false,
			expectNil:     false,
		},
		{
			name:  "no rows returns nil",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id = $1",
			args:  []any{999},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"})
				mock.ExpectQuery("SELECT (.+)").WithArgs(999).WillReturnRows(rows)
			},
			expectedError: false,
			expectNil:     true,
		},
		{
			name:  "query error",
			query: "SELECT id, first_name, last_name, email FROM users",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+)").WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
			expectNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			result, err := SelectSingle[TestUser](tx, tt.query, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectNil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSelect(t *testing.T) {
	Register[TestUser](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	tests := []struct {
		name          string
		query         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		expectedCount int
	}{
		{
			name:  "multiple rows",
			query: "SELECT id, first_name, last_name, email FROM users",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).
					AddRow(1, "John", "Doe", "john@example.com").
					AddRow(2, "Jane", "Smith", "jane@example.com")
				mock.ExpectQuery("SELECT (.+)").WillReturnRows(rows)
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:  "empty result set",
			query: "SELECT id, first_name, last_name, email FROM users WHERE id > $1",
			args:  []any{1000},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"})
				mock.ExpectQuery("SELECT (.+)").WithArgs(1000).WillReturnRows(rows)
			},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:  "query error",
			query: "SELECT id, first_name, last_name, email FROM users",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+)").WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			result, err := Select[TestUser](tx, tt.query, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestInsert(t *testing.T) {
	Register[TestUser](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	tests := []struct {
		name          string
		user          *TestUser
		setupMock     func(sqlmock.Sqlmock)
		expectedID    int
		expectedError bool
	}{
		{
			name: "successful insert",
			user: &TestUser{FirstName: "John", LastName: "Doe", Email: "john@example.com"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO (.+) RETURNING id").WithArgs("John", "Doe", "john@example.com").WillReturnResult(sqlmock.NewResult(42, 1))
			},
			expectedID:    42,
			expectedError: false,
		},
		{
			name: "query error",
			user: &TestUser{FirstName: "John", LastName: "Doe", Email: "john@example.com"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO (.+)").WillReturnError(sql.ErrTxDone)
			},
			expectedID:    0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			id, err := Insert[TestUser](tx, tt.user)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestInsertUuid(t *testing.T) {
	Register[TestUuidEntity](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	entity := &TestUuidEntity{Name: "Test", Description: "Description"}

	db, mock, tx := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec("INSERT INTO (.+)").
		WillReturnResult(sqlmock.NewResult(0, 1))

	uuidStr, err := InsertUuid[TestUuidEntity](tx, entity)

	assert.NoError(t, err)
	assert.NotEmpty(t, uuidStr)

	// Verify UUID is valid format
	_, err = uuid.Parse(uuidStr)
	assert.NoError(t, err)

	// Verify UUID was set on entity
	assert.Equal(t, uuidStr, entity.Id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertExistingUuid(t *testing.T) {
	Register[TestUuidEntity](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	existingUuid := uuid.New().String()
	entity := &TestUuidEntity{Id: existingUuid, Name: "Test", Description: "Description"}

	db, mock, tx := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec("INSERT INTO (.+)").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := InsertExistingUuid[TestUuidEntity](tx, entity)

	assert.NoError(t, err)
	assert.Equal(t, existingUuid, entity.Id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate(t *testing.T) {
	Register[TestUser](DefaultDbNamingStrategy{}, mockQueryGenerator{})

	tests := []struct {
		name          string
		user          *TestUser
		where         string
		args          []any
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		errorContains string
	}{
		{
			name:  "successful update",
			user:  &TestUser{Id: 1, FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"},
			where: "id = $1",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE test_users SET id = \\$1,first_name = \\$2,last_name = \\$3,email = \\$4 WHERE id = \\$1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:          "empty where clause",
			user:          &TestUser{Id: 1, FirstName: "Jane"},
			where:         "",
			args:          []any{},
			setupMock:     func(mock sqlmock.Sqlmock) {},
			expectedError: true,
			errorContains: "parameter 'where' was not present",
		},
		{
			name:  "exec error",
			user:  &TestUser{Id: 1, FirstName: "Jane"},
			where: "id = $1",
			args:  []any{1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE (.+)").WillReturnError(sql.ErrTxDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, tx := setupMockDB(t)
			defer db.Close()

			tt.setupMock(mock)

			err := Update[TestUser](tx, tt.user, tt.where, tt.args...)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if !tt.expectedError {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
