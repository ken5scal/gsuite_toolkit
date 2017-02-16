package main

import (
	"log"

	"github.com/ken5scal/gsuite_toolkit/client"

	"encoding/csv"
	"fmt"
	//admin "google.golang.org/api/admin/directory/v1"
	report "google.golang.org/api/admin/reports/v1"
	"io"
	"os"
	"strings"
	"github.com/ken5scal/gsuite_toolkit/services/reports"
	"github.com/urfave/cli"
	"sort"
)

const (
	clientSecretFileName = "client_secret.json"
	subCommandReport = "report"
)

func main() {
	app := cli.NewApp()
	app.Name = "gsuite"
	app.Usage = "help managing gsuite"
	app.Version = "0.1"
	app.Authors = []cli.Author{{Name: "Kengo Suzuki", Email:"kengoscal@gmai.com"}}
	app.Action  = func(c *cli.Context) error {
		if c.NArg() == 0 {
			cli.ShowAppHelp(c)
		}
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name: subCommandReport,
			Category: subCommandReport,
			Subcommands: []cli.Command{
				{
					Name:  "2sv",
					Usage: "get employees who have not enabled 2sv",
					Action: GetReportNon2StepVerifiedUsers,
				},
				{
					Name:  "remove",
					Usage: "remove an existing template",
					Action: func(c *cli.Context) error {
						fmt.Println("removed task template: ", c.Args().First())
						return nil
					},
				},
			},
		},
		{
			Name: "login",
			Category: subCommandReport,
		},
	}

	//app.Before = altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("flagfilename"))
	//app.Flags = flags

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)

	//scopes := []string{
	//	admin.AdminDirectoryOrgunitScope, admin.AdminDirectoryUserScope,
	//	report.AdminReportsAuditReadonlyScope, report.AdminReportsUsageReadonlyScope,
	//}
	//c := client.NewClient(clientSecretFileName, scopes)
	//goneUsers, err := users.GetUsersWhoHasNotLoggedInFor30Days(c.Client)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//for _, user := range goneUsers {
	//	fmt.Println(user.PrimaryEmail)
	//}
	//
	//
	//s, err := reports.NewService(c.Client)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//loginData, _ := s.GetEmployeesNotLogInFromOfficeIP()
	//
	//for key, value := range loginData {
	//	if !value.OfficeLogin {
	//		fmt.Println(key)
	//		fmt.Print("     IP: ")
	//		fmt.Println(value.LoginIPs)
	//	}
	//}

	//GetReportNon2StepVerifiedUsers(err, s)
	//for _, user := range non2SVuser.Users {
	//	fmt.Println(user.Entity.UserEmail)
	//}

	//
	//payload := constructPayload("/non2SVuser/suzuki/Desktop/org_structure.csv")
	//fmt.Println(payload)
	//url := "https://www.googleapis.com/batch"
	//
	//req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
	//req.Header.Add("content-type", "multipart/mixed; boundary=batch_0123456789")
	//req.Header.Add("authorization", "Bearer someToken")
	//res, _ := c.Do(req)
	//
	//defer res.Body.Close()
	//_, err = ioutil.ReadAll(res.Body)
	//if err != nil {
	//	log.Fatalln(err)
	//}
}

func GetReportNon2StepVerifiedUsers(context *cli.Context) error {
	client := client.NewClient(clientSecretFileName, []string{report.AdminReportsUsageReadonlyScope})
	s, err := reports.NewService(client)
	if err != nil {
		return err
	}
	non2svUserReports, err := s.GetNon2StepVerifiedUsers()
	if err != nil {
		return err
	}

	fmt.Println("Latest Report: " + non2svUserReports.TimeStamp.String())
	for _, user := range non2svUserReports.Users {
		fmt.Println(user.Entity.UserEmail)
	}
	return nil
}

func constructPayload(filePath string) string {
	var reader *csv.Reader
	var row []string
	var payload string
	boundary := "batch_0123456789"
	header := "--" + boundary + "\nContent-Type: application/http\n\n"

	csv_file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer csv_file.Close()
	reader = csv.NewReader(csv_file)

	for {
		row, err = reader.Read()
		if err == io.EOF {
			return payload + "--batch_0123456789--"
		}

		if strings.Contains(row[5], "@moneyforward.co.jp") && !strings.Contains(payload, row[5]) {
			payload = payload + header + RequestLine("PUT", row[5]) + "\n\n"
		}
	}
}

func RequestLine(method string, email string) string {
	//return "GET https://www.googleapis.com/admin/directory/v1/users/" +  email
	return method + " " + "https://www.googleapis.com/admin/directory/v1/users/" + email + "\n" +
		"Content-Type: application/json\n\n" + Body()
}

func Body() string {
	return "{\n" + "\"orgUnitPath\": \"/社員・委託社員・派遣社員・アルバイト\"\n" + "}\n"
}
