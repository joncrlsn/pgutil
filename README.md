pgutil
======

GoLang PostgreSQL utilities I've written and collected from various public sources on the web.

```
// Database connection info
type DbInfo struct {
    DbName string
    DbHost string
    DbPort int32
    DbUser string
    DbPass string
    DbOptions string
}
```

```
/*
 * Populates the database connection info from environment variables and/or runtime flags.
 * flag.Parse() is called, so define any other program flags before calling this.
 */
func (dbInfo *DbInfo) Populate() {
```

```
/*
 * Opens a postgreSQL database connection using the DbInfo instance
 */
func (dbInfo *DbInfo) Open() (*sql.DB, error)
```
