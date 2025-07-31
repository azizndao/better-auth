# Custom Database Configuration

Better Auth now supports custom GORM dialectors and configurations, allowing you to use any database supported by GORM (PostgreSQL, MySQL, SQLite, SQL Server, etc.) with your own connection parameters and GORM options.

## Benefits

- **Flexibility**: Use any GORM-supported database driver
- **Customization**: Pass custom GORM configurations (logging, connection pooling, etc.)
- **No Driver Dependencies**: Better Auth doesn't include database drivers - you choose what you need
- **Backwards Compatibility**: Default SQLite behavior is preserved

## Usage

### Basic Usage (Default SQLite)

```go
// Uses default SQLite database (./better-auth.db)
auth, err := betterauth.New(cfg)
```

### Custom Database Configuration

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// Example 1: PostgreSQL with custom configuration
dialector := postgres.Open("host=localhost user=username password=password dbname=mydb port=5432 sslmode=disable")
gormConfig := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
}
dbConfig := betterauth.NewDatabaseConfig(dialector, gormConfig)
auth, err := betterauth.NewWithDatabase(cfg, dbConfig)

// Example 2: MySQL with connection pooling
dialector := mysql.Open("user:password@tcp(localhost:3306)/better_auth?charset=utf8mb4&parseTime=True&loc=Local")
gormConfig := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Silent),
}
dbConfig := betterauth.NewDatabaseConfig(dialector, gormConfig)
auth, err := betterauth.NewWithDatabase(cfg, dbConfig)

// Example 3: SQLite with custom path
dialector := sqlite.Open("./custom-auth.db")
dbConfig := betterauth.NewDatabaseConfig(dialector, nil) // nil uses default GORM config
auth, err := betterauth.NewWithDatabase(cfg, dbConfig)
```

### With Plugin System

The custom database configuration works seamlessly with the plugin system:

```go
// Create custom database
dialector := postgres.Open("your-postgres-connection-string")
dbConfig := betterauth.NewDatabaseConfig(dialector, &gorm.Config{})
auth, err := betterauth.NewWithDatabase(cfg, dbConfig)
if err != nil {
    return err
}

// Register and enable plugins
jwtPlugin := jwt.NewJWTPlugin(&jwt.JWTConfig{...})
orgPlugin := organizations.NewOrganizationPlugin()

auth.RegisterPlugin(jwtPlugin)
auth.RegisterPlugin(orgPlugin)

// Plugin tables will be created in your custom database
auth.EnablePlugin(ctx, "jwt", jwtConfig)
auth.EnablePlugin(ctx, "organizations", orgConfig)
```

## API Reference

### `NewDatabaseConfig(dialector, config)`
Creates a new database configuration.

**Parameters:**
- `dialector` (gorm.Dialector): GORM dialector for your database
- `config` (*gorm.Config): GORM configuration options (can be nil for defaults)

**Returns:** `*database.DatabaseConfig`

### `NewWithDatabase(cfg, dbConfig)`
Creates a Better Auth instance with custom database configuration.

**Parameters:**
- `cfg` (*Config): Better Auth configuration
- `dbConfig` (*database.DatabaseConfig): Database configuration

**Returns:** `(*BetterAuth, error)`

## Supported Databases

Since this uses GORM dialectors, any database supported by GORM can be used:

- **PostgreSQL** - `gorm.io/driver/postgres`
- **MySQL** - `gorm.io/driver/mysql`
- **SQLite** - `gorm.io/driver/sqlite`
- **SQL Server** - `gorm.io/driver/sqlserver`
- **ClickHouse** - Third-party drivers available

## Migration

All plugin tables and core tables are automatically migrated when plugins are enabled. The conditional loading ensures only necessary tables are created based on enabled plugins.

## Examples

See the `/examples/custom-database/` directory for complete working examples with different database configurations.