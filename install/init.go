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

	sql := `CREATE TABLE IF NOT EXISTS Virtual (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                UID INT NOT NULL,
                Vname CHAR(20),
                Tag CHAR(20),
                Passwd CHAR(50),
                IPv4 CHAR(30),
                IPv6 CHAR(50),
                LocalIP CHAR(30),
                Vcpu INT NOT NULL,
                Vmemory INT NOT NULL,
                Status INT NOT NULL,
                Bandwidth INT NOT NULL,
                Br CHAR(10),
                Mac CHAR(20),
                Sys CHAR(20),
                ACpu INT NOT NULL,
                ABandwidth INT NOT NULL,
                AMemory INT NOT NULL,
                ADisk INT NOT NULL,
                AStatus INT NOT NULL,
                Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                Utime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                Etime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
            );
            CREATE TABLE IF NOT EXISTS Watch (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                Vname CHAR(20),
                CPU INT NOT NULL,
                Memory INT NOT NULL,
                Up INT NOT NULL,
                Down INT NOT NULL,
                Read INT NOT NULL,
                Write INT NOT NULL,
                Ctime INT NOT NULL
            );
            CREATE TABLE IF NOT EXISTS User (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                Username CHAR(20) NOT NULL,
                Passwd CHAR(50) NOT NULL,
                Tel CHAR(20) NOT NULL,
                Email CHAR(20) NOT NULL,
                Realname CHAR(20) NOT NULL,
                Idtype INT NOT NULL,
                Idnumber CHAR(20) NOT NULL,
                Sex INT NOT NULL,
                Company CHAR(20) NOT NULL,
                Address CHAR(20) NOT NULL,
                City CHAR(20) NOT NULL,
                Status INT NOT NULL,
                Utime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
            );
            CREATE TABLE IF NOT EXISTS Verify (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                Email CHAR(20) NOT NULL,
                Code CHAR(20) NOT NULL,
                Status INT NOT NULL,
                Vtime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
            );
            `
	db.Exec(sql)

	// `CREATE TABLE IF NOT EXISTS vv (
	//     ID INTEGER PRIMARY KEY AUTOINCREMENT,
	//     UID INT NOT NULL,
	//     Vname CHAR(20)
	// );
	//开机关机的时候做计费标记
}
