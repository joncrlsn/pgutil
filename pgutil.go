package pgutil

import "os"
import "flag"
import "log"
import "time"
import "fmt"
import "strconv"
import "strings"
import "database/sql"
import "github.com/joncrlsn/misc"
import "github.com/joncrlsn/fileutil"
import _ "github.com/lib/pq" // pg this import is for the PostgreSQL driver

var p = fmt.Println
var pgPassFile = ".pgpass"

const isoFormat = "2006-01-02T15:04:05.000-0700"

// DbInfo contains database connection info
type DbInfo struct {
	DbName    string
	DbHost    string
	DbPort    int32
	DbUser    string
	DbPass    string
	DbOptions string
}

// Populate populates the database connection info from environment variables or runtime flags.
// This calls flag.Parse(), so define any other program flags before calling this.
func (dbInfo *DbInfo) Populate() {
	hostDefault := misc.CoalesceStrings(os.Getenv("PGHOST"), "localhost")
	portDefaultStr := misc.CoalesceStrings(os.Getenv("PGPORT"), "5432")
	dbDefault := os.Getenv("PGDATABASE")
	userDefault := os.Getenv("PGUSER")
	passDefault := os.Getenv("PGPASSWORD")
	optionsDefault := os.Getenv("PGOPTIONS")

	// port is a little different because it's an int
	portDefault, _ := strconv.Atoi(portDefaultStr)
	fmt.Println("portDefault", portDefault)

	var dbUser = flag.String("U", userDefault, "db user")
	var dbPass = flag.String("pw", passDefault, "db password")
	var dbHost = flag.String("h", hostDefault, "db host")
	var dbPort = flag.Int("p", portDefault, "db port")
	var dbName = flag.String("d", dbDefault, "db name")
	var dbOptions = flag.String("o", optionsDefault, "db options (eg. sslmode=disable)")

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
		if *dbPass == "prompt" {
			dbInfo.DbPass = misc.PromptPassword("Enter password: ")
		} else if len(*dbPass) > 1 {
			dbInfo.DbPass = *dbPass
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
	if len(*dbOptions) > 0 {
		dbInfo.DbOptions = *dbOptions
	}
}

// ConnectionString returns the string needed by the postgres driver library to connect
func (dbInfo *DbInfo) ConnectionString() string {
	connString := "user=" + dbInfo.DbUser + " host=" + dbInfo.DbHost + " dbname=" + dbInfo.DbName + " password=" + dbInfo.DbPass
	if len(dbInfo.DbOptions) > 0 {
		connString += " " + dbInfo.DbOptions
	}
	return connString
}

// Open opens a postgreSQL database connection using the DbInfo instance
func (dbInfo *DbInfo) Open() (*sql.DB, error) {
	conn := dbInfo.ConnectionString()
	db, err := sql.Open("postgres", conn)
	return db, err
}

// DbUsage provides a model for adding to your own database executable
func DbUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-h <host>] [-p <port>] [-d <dbname>] [-U <user>] [-pw <password>] [-o <db option>] \n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

// PgPassword parses the ~/.pgpass file and gets the password for the given user.  The current implementation
// ignores the location field.
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

// QueryStrings returns row maps (keyed by the column name) in a channel.
// Dynamically converts each column value to a SQL string value.
// See http://stackoverflow.com/questions/23507531/is-golangs-sql-package-incapable-of-ad-hoc-exploratory-queries
func QueryStrings(db *sql.DB, query string) (chan map[string]string, []string) {
	rowChan := make(chan map[string]string)

	rows, err := db.Query(query)
	check("running query", err)
	columnNames, err := rows.Columns()
	check("getting column names", err)

	go func() {

		defer rows.Close()

		vals := make([]interface{}, len(columnNames))
		valPointers := make([]interface{}, len(columnNames))
		// Copy
		for i := 0; i < len(columnNames); i++ {
			valPointers[i] = &vals[i]
		}

		for rows.Next() {
			err = rows.Scan(valPointers...)
			check("scanning a row", err)

			row := make(map[string]string)
			// Convert each cell to a SQL-valid string representation
			for i, valPtr := range vals {
				//fmt.Println(reflect.TypeOf(valPtr))
				switch valueType := valPtr.(type) {
				case nil:
					row[columnNames[i]] = "null"
				case []uint8:
					row[columnNames[i]] = string(valPtr.([]byte))
				case string:
					row[columnNames[i]] = valPtr.(string)
				case int64:
					row[columnNames[i]] = fmt.Sprintf("%d", valPtr)
				case float64:
					row[columnNames[i]] = fmt.Sprintf("%f", valPtr)
				case bool:
					row[columnNames[i]] = fmt.Sprintf("%t", valPtr)
				case time.Time:
					row[columnNames[i]] = valPtr.(time.Time).Format(isoFormat)
				case fmt.Stringer:
					row[columnNames[i]] = fmt.Sprintf("%v", valPtr)
				default:
					row[columnNames[i]] = fmt.Sprintf("%v", valPtr)
					fmt.Println("Warning, column %s is an unhandled type: %v", columnNames[i], valueType)
				}
			}
			rowChan <- row
		}
		close(rowChan)
	}()
	return rowChan, columnNames
}

/*
 * Returns row maps (keyed by the column name) in a channel.
 * Dynamically converts each column value to a SQL string value.
 * See http://stackoverflow.com/questions/23507531/is-golangs-sql-package-incapable-of-ad-hoc-exploratory-queries
 */
//func QuerySalValues(db *sql.DB, query string) (chan map[string]string, []string) {
//	rowChan := make(chan map[string]string)
//
//	rows, err := db.Query(query)
//	check("running query", err)
//	columnNames, err := rows.Columns()
//	check("getting column names", err)
//
//	go func() {
//
//		defer rows.Close()
//
//		vals := make([]interface{}, len(columnNames))
//		valPointers := make([]interface{}, len(columnNames))
//		// Copy
//		for i := 0; i < len(columnNames); i++ {
//			valPointers[i] = &vals[i]
//		}
//
//		for rows.Next() {
//			err = rows.Scan(valPointers...)
//			check("scanning a row", err)
//
//			row := make(map[string]string)
//			// Convert each cell to a SQL-valid string representation
//			for i, valPtr := range vals {
//				//fmt.Println(reflect.TypeOf(valPtr))
//				switch valueType := valPtr.(type) {
//				case nil:
//					row[columnNames[i]] = "null"
//				case []uint8:
//					row[columnNames[i]] = "'" + string(valPtr.([]byte)) + "'"
//				case string:
//					row[columnNames[i]] = "'" + valPtr.(string) + "'"
//				case int64:
//					row[columnNames[i]] = fmt.Sprintf("%d", valPtr)
//				case float64:
//					row[columnNames[i]] = fmt.Sprintf("%f", valPtr)
//				case bool:
//					row[columnNames[i]] = fmt.Sprintf("%t", valPtr)
//				case time.Time:
//					row[columnNames[i]] = "'" + valPtr.(time.Time).Format(isoFormat) + "'"
//				case fmt.Stringer:
//					row[columnNames[i]] = fmt.Sprintf("%v", valPtr)
//				default:
//					row[columnNames[i]] = fmt.Sprintf("%v", valPtr)
//					fmt.Println("Column %s is an unhandled type: %v", columnNames[i], valueType)
//				}
//			}
//			rowChan <- row
//		}
//		close(rowChan)
//	}()
//	return rowChan, columnNames
//}

func check(msg string, err error) {
	if err != nil {
		log.Fatal("Error "+msg, err)
	}
}
