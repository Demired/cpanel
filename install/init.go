package install

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Init() {
	db, err := sql.Open("sqlite3", "./cpanel.db")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer db.Close()

	sql := `CREATE TABLE IF NOT EXISTS vvm (
                Id INTEGER PRIMARY KEY AUTOINCREMENT,
                UID INT NOT NULL,
                Vname CHAR(20),
                IPv4 CHAR(30),
                IPv6 CHAR(50),
                LocalIP CHAR(30),
                Vcpu INT NOT NULL,
                Vmemory INT NOT NULL,
                Status INT NOT NULL,
                Bandwidth INT NOT NULL,
                Mac CHAR(20),
                Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                Utime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
            );
            CREATE TABLE IF NOT EXISTS watch (
                Id INTEGER PRIMARY KEY AUTOINCREMENT,
                Vname CHAR(20),
                CPU INT NOT NULL,
                Memory INT NOT NULL,
                Ctime INT NOT NULL
            );
            // CREATE TABLE IF NOT EXISTS billing (
            //     ID INTEGER PRIMARY KEY AUTOINCREMENT,
            //     UID INT NOT NULL,
            //     VID INT NOT NULL,
            //     Status INT NOT NULL,
            //     Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            // );
            `
	db.Exec(sql)
	//开机关机的时候做计费标记
}
