package pgutil

import "os"
import "strings"
import "testing"
import "github.com/joncrlsn/fileutil"

func init() {
	// this override the value in pgutil.go
	pgPassFile = ".pgpass.testing"
	testFileName := os.Getenv("HOME") + "/" + pgPassFile
	lines := strings.Split(`
#hostname:port:database:username:password
*:*:*:c42:lKj*$hL;(~
*:*:*:c42ro:himom`, "\n")
	fileutil.WriteLinesSlice(lines, testFileName)
}

func Test_FindPgPassword(t *testing.T) {
	if len(PgPassword("c42")) == 10 {
		t.Log("find password test passed.")
	} else {
		t.Error("Correct password not found")
	}
}
