package pgutil

import "database/sql"
import _ "github.com/lib/pq"
import "os"
import "flag"
import "fmt"
import "strconv"
import "github.com/joncrlsn/misc"

var p = fmt.Println

// Database connection info
type DbInfo struct {
    DbName string
    DbHost string 
    DbPort int32
    DbUser string
    DbPass string
    DbOptions string
}

/*
 * Populates the database connection info from environment variables or runtime flags.
 * This calls flag.Parse(), so define any other program flags before calling this.
 */
func (dbInfo *DbInfo) Populate() {
    userDefault    := misc.CoalesceStrings(os.Getenv("DBUSER"), "c42ro")
    hostDefault    := misc.CoalesceStrings(os.Getenv("DBHOST"), "localhost")
    portDefaultStr := misc.CoalesceStrings(os.Getenv("DBPORT"), "5432")

    // port is a little different because it's an int
    portDefault, _ := strconv.Atoi(portDefaultStr)
    fmt.Println("portDefault", portDefault)

    var dbUser = flag.String("user", userDefault,  "db user")
    var dbPass = flag.String("pw",   "",           "db password")
    var dbHost = flag.String("host", hostDefault,  "db host")
    var dbPort = flag.Int("port",    portDefault,  "db port")
    var dbName = flag.String("db",   "",           "db name")

    // This will parse all the flags defined for the program.  Not sure how to get around this.
    flag.Parse()

    if len(*dbUser) > 0 {
        dbInfo.DbUser = *dbUser
    }
    if len(*dbPass) > 0 {
        dbInfo.DbPass = *dbPass
    }
    if len(*dbHost) > 0 {
        dbInfo.DbHost = *dbHost
    }
    if *dbPort > 0 {
        dbInfo.DbPort = int32(*dbPort)
    }
    if len(*dbName) > 0 {
        dbInfo.DbName = *dbName
    }
}

func (dbInfo *DbInfo) ConnectionString() string {
    connString := "user=" + dbInfo.DbUser + " host=" + dbInfo.DbHost + " dbname=" + dbInfo.DbName + " password=" + dbInfo.DbPass
    if len(dbInfo.DbOptions) > 0 {
        connString += " " + dbInfo.DbOptions
    }
    return connString 
}

/*
 * Opens a postgreSQL database connection using the DbInfo instance
 */
func (dbInfo *DbInfo) Open() (*sql.DB, error) {
    conn := dbInfo.ConnectionString()
	db, err := sql.Open("postgres", conn)
	return db, err
}

/* 
 * Provides a model for adding to your own database executable
 */
func DbUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-host <string>] [-port <int>] [-db <string>] [-user <string>] [-password <string>] \n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}
