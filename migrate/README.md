# Migrate

db migration based on [sql-migrate](https://github.com/rubenv/sql-migrate), only use migrate.Up.

Supports both **MySQL** and **PostgreSQL** databases.

## Usage

### sql files

prepare sql files before migrate

```bash
mkdir db

cat <<EOF > db/2022120710-user.sql
-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE user
(
	id       BIGINT UNSIGNED AUTO_INCREMENT COMMENT 'auto increment id' PRIMARY KEY,
	username VARCHAR(191) NULL COMMENT 'user login name'
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci;
EOF


cat <<EOF > db/2022120711-role.sql
-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE role
(
	id   BIGINT UNSIGNED AUTO_INCREMENT COMMENT 'auto increment id' PRIMARY KEY,
	name VARCHAR(50) NULL COMMENT 'name'
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci;
EOF
```

> caution: must set file header!  
> -- +migrate Up  
> -- SQL in section 'Up' is executed when this migration is applied

### Do

```bash
go get -u github.com/go-cinch/common/migrate
```

#### MySQL Example

```go
import (
	"embed"
	"fmt"
	"github.com/go-cinch/common/migrate"
)

//go:embed db
var db embed.FS

func main() {
	err := migrate.Do(
		migrate.WithDriver("mysql"),
		migrate.WithUri("root:root@tcp(127.0.0.1:3306)/test?parseTime=true"),
		migrate.WithFs(db),
		migrate.WithFsRoot("db"),
	)
	fmt.Println(err)
}
```

#### PostgreSQL Example

```go
import (
	"embed"
	"fmt"
	"github.com/go-cinch/common/migrate"
)

//go:embed db
var db embed.FS

func main() {
	err := migrate.Do(
		migrate.WithDriver("postgres"),
		migrate.WithUri("host=localhost user=postgres password=postgres dbname=test port=5432 sslmode=disable"),
		migrate.WithFs(db),
		migrate.WithFsRoot("db"),
	)
	fmt.Println(err)
}
```

Or using PostgreSQL URL format:

```go
err := migrate.Do(
	migrate.WithDriver("postgres"),
	migrate.WithUri("postgres://postgres:postgres@localhost:5432/test?sslmode=disable"),
	migrate.WithFs(db),
	migrate.WithFsRoot("db"),
)
```

## Options

- `WithCtx` - context
- `WithDriver` - database driver, supports `mysql` and `postgres`, default `mysql`
- `WithUri` - database uri, if database not exist will auto create
  - MySQL: `root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&timeout=10000ms`
  - PostgreSQL (key=value format): `host=localhost user=postgres password=postgres dbname=test port=5432 sslmode=disable`
  - PostgreSQL (URL format): `postgres://postgres:postgres@localhost:5432/test?sslmode=disable`
- `WithChangeTable` - change history table name, default `schema_migrations`
- `WithBefore` - callback function, custom callback before exec sql script
- `WithFs` - embed files
- `WithFsRoot` - embed root path

## Features

- **Auto Database Creation**: Automatically creates the database if it doesn't exist
- **Migration Status**: Shows pending and applied migrations before execution
- **Rollback Support**: Supports rollback via `SQL_MIGRATE_ROLLBACK` environment variable
