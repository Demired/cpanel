package install

import(
        "fmt"
        "database/sql"
        _ "github.com/mattn/go-sqlite3"
)

func Init(){
        db,err := sql.Open("sqlite3","./cpanel.db")
        if err != nil {
            fmt.Println(err.Error())
            return
        }

        defer db.Close()

sql := `
CREATE TABLE IF NOT EXISTS vm (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    Vname CHAR(20),
    IPv4 CHAR(30),
    IPv6 CHAR(50),
    LocalIP CHAR(30),
    Vcpu INT NOT NULL,
    Vmemory INT NOT NULL,
    Status INT NOT NULL,
    Ctime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    Utime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
`
        db.Exec(sql)

    db.Query("INSERT INTO vm (Vname,IPv4,IPv6,LocalIP,Vcpu,Vmemory,Status) VALUES ('zhangyuan','123.123.123.123','asd1:12asd:ss454:123a:f2a','127.0.0.1',1,2,1);")
	db.Query("INSERT INTO vm (Vname,IPv4,IPv6,LocalIP,Vcpu,Vmemory,Status) VALUES ('0x8c','123.123.124.124','12asd:ss454:123a:f2a:asd1','127.0.0.2',2,1,1);")
}

