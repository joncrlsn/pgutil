package pgutil

import "database/sql"
import _ "github.com/lib/pq"
import "os"
import "flag"
import "fmt"
import "strconv"
import "strings"
import "github.com/joncrlsn/misc"
import "github.com/joncrlsn/fileutil"

var p = fmt.Println
var pgPassFile string = ".pgpass"

// Database connection info
type DbInfo struct {
	DbName    string
	DbHost    string
	DbPort    int32
	DbUser    string
	DbPass    string
	DbOptions string
}

/*
 * Populates the database connection info from environment variables or runtime flags.
 * This calls flag.Parse(), so define any other program flags before calling this.
 */
func (dbInfo *DbInfo) Populate() {
	userDefault := os.Getenv("DBUSER")
	hostDefault := misc.CoalesceStrings(os.Getenv("DBHOST"), "localhost")
	portDefaultStr := misc.CoalesceStrings(os.Getenv("DBPORT"), "5432")
	passDefault := os.Getenv("PGPASS")

	// port is a little different because it's an int
	portDefault, _ := strconv.Atoi(portDefaultStr)
	fmt.Println("portDefault", portDefault)

	var dbUser = flag.String("U", userDefault, "db user")
	var dbPass = flag.String("pw", "", "db password")
	var dbHost = flag.String("h", hostDefault, "db host")
	var dbPort = flag.Int("p", portDefault, "db port")
	var dbName = flag.String("d", "", "db name")

	// This will parse all the flags defined for the program.  Not sure how to get around this.
	flag.Parse()

	if len(*dbUser) > 0 {
		dbInfo.DbUser = *dbUser
	}
	if len(*dbPass) > 0 {
		dbInfo.DbPass = *dbPass
	}
	// the password is a little different because it can also be found in ~/.pgpass
	if len(dbInfo.DbPass) == 0 {
		if len(passDefault) > 1 {
			dbInfo.DbPass = passDefault
		} else {
			dbInfo.DbPass = PgPassword(dbInfo.DbUser)
            if len(dbInfo.DbPass) == 0 {
                dbInfo.DbPass = misc.PromptPassword("Enter password: ")
            }
		}
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
	fmt.Fprintf(os.Stderr, "usage: %s [-h <host>] [-p <port>] [-d <dbname>] [-U <user>] [-pw <password>] \n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

/*
 * Parses the ~/.pgpass file and gets the password for the given user.  The current implementation
 * ignores the location field.
 */
func PgPassword(user string) string {
	pgPassPath := os.Getenv("HOME") + "/" + pgPassFile
	exists, err := fileutil.Exists(pgPassPath)
	if err != nil {
		panic(err)
	}
	if !exists {
		return ""
	}

	lines, err := fileutil.ReadLinesArray(pgPassPath)
	if err != nil {
		panic(err)
	}
	for _, line := range lines {
		if strings.Contains(line, ":"+user+":") {
			fields := strings.Split(line, ":")
			password := fields[4]
            fmt.Println("Used password from ~/.pgpass")
			return password
		}
	}
	return ""
}
