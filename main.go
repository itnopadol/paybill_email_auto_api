package main

import (
	"github.com/jmoiron/sqlx"
	"fmt"
	"net/http"
	"log"
	"github.com/vigneshuvi/GoDateFormat"
	"time"
	"github.com/satori/go.uuid"
)

type Customer struct {
	ArCode string `json:"ar_code" db:"ArCode"`
	ArName string `json:"ar_name" db:"ArName"`
	DocNo  string `json:"doc_no" db:"DocNo"`
	EmailAddress  string `json:"email_address" db:"EmailAddress"`
}

var dbc *sqlx.DB

func init() {
	dbc = ConnectSQL()
}

func ConnectSQL() (msdb *sqlx.DB) {
	db_host := "192.168.0.7"
	db_name := "bcnp"
	db_user := "sa"
	db_pass := "[ibdkifu"
	port := "1433"

	dsn := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", db_host, db_user, db_pass, port, db_name)
	msdb = sqlx.MustConnect("mssql", dsn)
	if msdb.Ping() != nil {
		fmt.Println("Error")
	}

	fmt.Println("msdb = ", msdb.DriverName())
	return msdb
}

func GetToday(format string) (todayString string){
	today := time.Now()
	todayString = today.Format(format);
	return
}


func main() {

	fmt.Println((GetToday(GoDateFormat.ConvertFormat("HH:MM"))))

	timeChan := time.NewTimer(time.Second).C

	tickChan := time.NewTicker(time.Millisecond * 400).C

	doneChan := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond )
		fmt.Println("Now = ",(GetToday(GoDateFormat.ConvertFormat("HH"))),(GetToday(GoDateFormat.ConvertFormat("MM"))))

	}()

	for {
		select {
		case <- timeChan:
			fmt.Println("Timer expired")
		case <- tickChan:
			fmt.Println("Ticker ticked")
			if (GetToday(GoDateFormat.ConvertFormat("HH"))=="10" && GetToday(GoDateFormat.ConvertFormat("MM"))=="56") {
				custs := []Customer{}

				sql := `set dateformat dmy   select	 top 10 a.ArCode,b.name1 as ArName,a.DocNo,'it@nopadol.com' as EmailAddress from	dbo.bcpaybill a inner join dbo.bcar b on a.arcode = b.code where	cast(rtrim(day(a.createdatetime))+'/'+rtrim(month(a.createdatetime))+'/'+rtrim(year(a.createdatetime)) as datetime) = '14/02/2018'and a.docno not in (select distinct isnull(docno,'') as docno from npmaster.dbo.TB_CD_PaybillLogs) and isnull(b.emailaddress,'') <> '' and a.billstatus <> 1`
				err := dbc.Select(&custs, sql)
				if err != nil {
					fmt.Println(err.Error())
				}

				for _, c := range custs {
					uuid_token := uuid.Must(uuid.NewV4()).String()
					fmt.Printf("UUIDv4-1: %s\n", uuid_token)
					fmt.Println("ArName = ",c.ArCode, c.ArName)
					sql_insert := `Insert into NPMaster.dbo.TB_CD_PaybillLogs(SendDate,ArCode,DocNo,EmailAddress,SendDateTime,AccessToken) values(cast(rtrim(day(getdate()))+'/'+rtrim(month(getdate()))+'/'+rtrim(year(getdate())) as datetime),?,?,?,getdate(),?)`
					fmt.Println("sql_insert =",sql_insert, c.ArCode,c.DocNo, c.EmailAddress, uuid_token)
					dbc.Exec(sql_insert, c.ArCode,c.DocNo, c.EmailAddress, uuid_token)

					//xurl := "http://venus:8099/email?ar_code="+c.ArCode+"&doc_no="+c.DocNo+"&email="+c.EmailAddress+"&access_token="+uuid_token
					xurl := "http://localhost:8099/email?ar_code="+c.ArCode+"&ar_name="+c.ArName+"&doc_no="+c.DocNo+"&email="+c.EmailAddress+"&access_token="+uuid_token

					url := xurl
					fmt.Println("url = ", url)

					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						log.Fatal("NewRequest: ", err)
					}

					client := &http.Client{}

					resp, err := client.Do(req)
					if err != nil {
						log.Fatal("Do: ", err)
					}

					defer resp.Body.Close()

				}
				doneChan <- true
			}
		case <- doneChan:
			fmt.Println("Done")
			return
		}
	}

}
