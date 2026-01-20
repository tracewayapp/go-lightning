# OUTDATED - NEEDS UPDATE

<div align="center">
  <img width="150" height="200" alt="Untitled copy" src="https://github.com/user-attachments/assets/991b8172-9413-4f52-9cd8-6acef7bc042b" />
</div>

# go-lightning

`go-lightning` is a lightweight, high-performance database interaction library for Go. It is designed to be slim, fast, and easy to use, especially when working with projections and Data Transfer Objects (DTOs).

The project is currently used in a **production environment**.

## Key Features

- **Unified API**: The `lit` package provides a single API for both PostgreSQL and MySQL, with driver-specific optimizations handled internally.
- **Lightweight Projections**: The biggest advantage of `go-lightning` is its ability to load DTOs and projections with minimal effort. Regardless of your table structure, mapping a query result to a Go struct is straightforward and clean.
- **MySQL and PostgreSQL Support**: Register your models with the appropriate driver and the library handles query generation and driver-specific optimizations.
- **Generic CRUD Operations**: Automatic generation of `INSERT` and `UPDATE` queries for registered types.
- **Works with DB and Tx**: All operations accept both `*sql.DB` and `*sql.Tx` via the `Executor` interface.
- **Minimal Dependencies**: Keeps your project slim and focused.

## Docs

Documentation is available at https://tracewayapp.github.io/go-lightning

## Usage Limitations

- **ID Column Requirement**: For automatic `InsertGeneric` and `UpdateGeneric` operations, the library requires your tables to have an `id` column. It does not support tables without a primary `id` field for these specific automatic operations.

## Installation

```bash
go get github.com/tracewayapp/go-lightning/lit
```

## Configuration & Usage

### 1. Registration

Every model you intend to use with generic functions must be registered with a specific driver.

```go
import "github.com/tracewayapp/go-lightning/lit"

type User struct {
    Id        int
    FirstName string
    LastName  string
    Email     string
}

func init() {
    // Register for PostgreSQL
    lit.RegisterModel[User](lit.PostgreSQL)
    // OR Register for MySQL
    lit.RegisterModel[User](lit.MySQL)
}
```

### 2. PostgreSQL Usage

```go
import (
    "github.com/tracewayapp/go-lightning/lit"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func example(db *sql.DB) {
    // Insert - returns auto-generated ID
    id, _ := lit.InsertGeneric(db, &User{FirstName: "Jane", LastName: "Smith"})

    // Select Single
    user, _ := lit.SelectGenericSingle[User](db, "SELECT * FROM users WHERE id = $1", id)

    // Select Multiple
    users, _ := lit.SelectGeneric[User](db, "SELECT * FROM users WHERE last_name = $1", "Smith")

    // Update
    user.Email = "jane@example.com"
    _ = lit.UpdateGeneric(db, user, "id = $1", user.Id)

    // Delete
    _ = lit.Delete(db, "DELETE FROM users WHERE id = $1", user.Id)
}
```

### 3. MySQL Usage

```go
import (
    "github.com/tracewayapp/go-lightning/lit"
    _ "github.com/go-sql-driver/mysql"
)

func example(db *sql.DB) {
    // Insert
    id, _ := lit.InsertGeneric(db, &User{FirstName: "John", LastName: "Doe"})

    // Select Single
    user, _ := lit.SelectGenericSingle[User](db, "SELECT * FROM users WHERE id = ?", id)

    // Select Multiple
    users, _ := lit.SelectGeneric[User](db, "SELECT * FROM users WHERE last_name = ?", "Doe")

    // Update
    user.Email = "john@example.com"
    _ = lit.UpdateGeneric(db, user, "id = ?", user.Id)

    // Delete
    _ = lit.Delete(db, "DELETE FROM users WHERE id = ?", user.Id)
}
```

### 4. Working with Transactions

All operations work with both `*sql.DB` and `*sql.Tx`:

```go
func exampleWithTx(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // All operations accept tx
    id, err := lit.InsertGeneric(tx, &User{FirstName: "John"})
    if err != nil {
        return err
    }

    user, err := lit.SelectGenericSingle[User](tx, "SELECT * FROM users WHERE id = $1", id)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

### 5. UUID Support

For models with string ID fields, use UUID-specific insert functions:

```go
type Product struct {
    Id    string
    Name  string
    Price int
}

func init() {
    lit.RegisterModel[Product](lit.PostgreSQL)
}

func example(db *sql.DB) {
    // Auto-generate UUID
    uuid, _ := lit.InsertGenericUuid(db, &Product{Name: "Widget", Price: 100})

    // Use existing UUID
    product := &Product{Id: "my-custom-uuid", Name: "Gadget", Price: 200}
    _ = lit.InsertGenericExistingUuid(db, product)
}
```

### 6. Helper Functions

```go
// JoinForIn - for integer IN clauses
ids := []int{1, 2, 3}
query := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s)", lit.JoinForIn(ids))

// JoinStringForIn - generates driver-appropriate placeholders
// PostgreSQL: $1,$2,$3 (with offset support)
// MySQL: ?,?,?
names := []string{"a", "b", "c"}
placeholders := lit.JoinStringForIn[User](0, names)

// Or specify driver explicitly
placeholders := lit.JoinStringForInWithDriver(lit.PostgreSQL, 0, 3) // "$1,$2,$3"
placeholders := lit.JoinStringForInWithDriver(lit.MySQL, 0, 3)      // "?,?,?"
```

## Contributions

We welcome all contributions to the go-lightning project. You can open issues or PR and we will review and promptly merge them.

## Roadmap

- [ ] Add support for ClickHouse
- [ ] Add more examples
- [ ] Add support for named fields `db:"column_name"` (this can be done by using a naming strategy currently)
- [ ] Add a project homepage
- [ ] Add project docs
- [ ] Add support for composite primary keys
- [ ] Escaping SQL keywords for field names and table names

## Project Philosophy

- **Developer Written**: All core logic and architectural decisions were made and implemented by an actual developer.
- **AI Assisted Testing**: AI was utilized to help generate a comprehensive test suite as well as help out with documentation.

## License

MIT
