<div align="center">
  <img width="150" height="200" alt="Untitled copy" src="https://github.com/user-attachments/assets/991b8172-9413-4f52-9cd8-6acef7bc042b" />
</div>

# go-lightning

`go-lightning` is a lightweight, high-performance database interaction library for Go. It is designed to be slim, fast, and easy to use, especially when working with projections and Data Transfer Objects (DTOs).

The project is currently used in a **production environment**.

## Key Features

- **Lightweight Projections**: The biggest advantage of `go-lightning` is its ability to load DTOs and projections with minimal effort. Regardless of your table structure, mapping a query result to a Go struct is straightforward and clean.
- **MySQL and PostgreSQL Support**: Specialized modules `lmy` (MySQL) and `lpg` (PostgreSQL) provide tailored query generation and driver-specific optimizations.
- **Generic CRUD Operations**: Automatic generation of `INSERT` and `UPDATE` queries for registered types.
- **Minimal Dependencies**: Keeps your project slim and focused.

## Docs

Documentation is available at https://tracewayapp.github.io/go-lightning

## Usage Limitations

- **ID Column Requirement**: For automatic `InsertGeneric` and `UpdateGeneric` operations, the library requires your tables to have an `id` column. It does not support tables without a primary `id` field for these specific automatic operations.

## Installation MySQL

```bash
go get github.com/tracewayapp/go-lightning/lmy
```

## Installation PostgreSQL

```bash
go get github.com/tracewayapp/go-lightning/lpg
```

## Configuration & Usage

### 1. Registration

Every model you intend to use with generic functions must be registered. This sets up the internal field mapping and query generation.

```go
type User struct {
    Id        int
    FirstName string
    LastName  string
    Email     string
}

func init() {
    // For MySQL
    lmy.Register[User]()
    // OR For PostgreSQL
    lpg.Register[User]()
}
```

### 2. MySQL (using `lmy`)

To use `go-lightning` with MySQL, use the `lmy` package.

```go
import (
    "github.com/tracewayapp/go-lightning/lmy"
    _ "github.com/go-sql-driver/mysql"
)

func example(tx *sql.Tx) {
    // Insert
    id, _ := lmy.InsertGeneric(tx, &User{FirstName: "John", LastName: "Doe"})

    // Select Single
    user, _ := lmy.SelectGenericSingle[User](tx, "SELECT * FROM users WHERE id = ?", id)

    // Select Multiple
    users, _ := lmy.SelectGeneric[User](tx, "SELECT * FROM users WHERE last_name = ?", "Doe")
}
```

### 3. PostgreSQL (using `lpg`)

To use `go-lightning` with PostgreSQL, use the `lpg` package. It correctly handles the `$n` parameter syntax and `RETURNING id` logic.

```go
import (
    "github.com/tracewayapp/go-lightning/lpg"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func example(tx *sql.Tx) {
    // Insert
    id, _ := lpg.InsertGeneric(tx, &User{FirstName: "Jane", LastName: "Smith"})

    // Select Single
    user, _ := lpg.SelectGenericSingle[User](tx, "SELECT * FROM users WHERE id = $1", id)

    // Select Multiple
    users, _ := lpg.SelectGeneric[User](tx, "SELECT * FROM users WHERE last_name = $1", "Smith")
}
```

## Contributions

We welcome all contributions to the go-lightning project. You can open issues or PR and we will review and promptly merge them.

## Roadmap

- [ ] Add support for clickhouse
- [ ] Add more examples
- [ ] Add support for named fields `db:"column_name"` (this can be done by using a naming strategy currently)
- [ ] Add a project homepage
- [ ] Add project docs
- [ ] Add support for composite primary keys
- [ ] Escaping sql keywords for field names and table names

## Project Philosophy

- **Developer Written**: All core logic and architectural decisions were made and implemented by an actual developer.
- **AI Assisted Testing**: AI was utilized to help generate a comprehensive test suite as well as help out with documentation.

## License

MIT
