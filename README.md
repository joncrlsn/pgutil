pgutil
======

GoLang PostgreSQL utilities I've written

/*
 * Populates the database connection info from environment variables or runtime flags.
 * This calls flag.Parse(), so define any other program flags before calling this.
 */
func (dbInfo *DbInfo) Populate() {

/*
 * Opens a postgreSQL database connection using the DbInfo instance
 */
func (dbInfo *DbInfo) Open() (*sql.DB, error)


