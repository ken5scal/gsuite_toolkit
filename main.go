package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ken5scal/gsuite_toolkit/actions"
	"github.com/ken5scal/gsuite_toolkit/client"
	"github.com/ken5scal/gsuite_toolkit/models"
	"github.com/urfave/cli"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"github.com/ken5scal/gsuite_toolkit/services"
)

const (
	ClientSecretFileName = "client_secret.json"
	CommandLogin         = "login"
)

type network struct {
	Name string
	Ip   []string
}

func main() {
	var tomlConf models.TomlConfig
	var s services.Service
	var a actions.Action
	var gsuiteClient *http.Client

	_, err := toml.DecodeFile("gsuite_config.toml", &tomlConf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	showHelpFunc := func(c *cli.Context) error {
		if c.NArg() == 0 {
			cli.ShowAppHelp(c)
		}
		return nil
	}

	setServiceToAction := func(s services.Service, a actions.Action) error {
		if err := a.SetService(s); err != nil {
			return err
		}
		return nil
	}

	app := cli.NewApp()
	app.Name = "gsuite"
	app.Usage = "help managing gsuite"
	app.Version = "0.1"
	app.Authors = []cli.Author{{Name: "Kengo Suzuki", Email: "kengoscal@gmai.com"}}
	app.Action = showHelpFunc

	gsuiteClient, err = client.CreateConfig().
		SetClientSecretFilename(ClientSecretFileName).
		SetScopes(tomlConf.Scopes).
		Build()
	if err != nil {
		fmt.Errorf("Failed building client: %v", err)
		return
	}
	app.Commands = []cli.Command{
		{
			Name: "group", Category: "group",
			Usage: "Audit and manage groups within GSuite",
			Before: func(context *cli.Context) error {
				s = services.InitGroupService()
				if err = s.SetClient(gsuiteClient); err != nil {
					return nil
				}
				a = actions.InitGroupAction()
				return setServiceToAction(s, a)
			},
			Action: showHelpFunc,
			Subcommands: []cli.Command {
				{
					Name: "list", Usage: "list existing groups",
					Action: func(context *cli.Context) error {
						return a.(*actions.GroupAction).RetrieveAllGroups()
					},
				},
			},
		},
		{
			Name: "drive", Category: "drive",
			Usage: "Audit files within Google Drive",
			Before: func(*cli.Context) error {
				s = services.InitDriveService()
				if err = s.SetClient(gsuiteClient); err != nil {
					return nil
				}
				a = actions.InitDriveAction()
				return setServiceToAction(s, a)
			},
			Action: showHelpFunc,
			Subcommands: []cli.Command{
				{
					Name: "list", Usage: "list existing files",
					Action: func(context *cli.Context) error {
						return a.(*actions.DriveAction).SearchAllFolders()
					},
				},
				{
					Name: "search", Usage: "search a keyword buy specifying an argument",
					Action: func(context *cli.Context) error {
						if context.NArg() != 1 {
							return errors.New("Number of keyword must be exactly 1")
						}
						return a.(*actions.DriveAction).SearchFoldersByName(context.Args()[0])
					},
				},
			},
		},
		{
			Name: CommandLogin, Category: CommandLogin, Usage: "Gain insights on content management with Google Drive activity reports. Audit administrator actions. Generate customer and user usage reports.",
			Before: func(*cli.Context) error {
				a = actions.InitLoginAction()
				s = services.InitReportService()
				if err = s.SetClient(gsuiteClient); err != nil {
					return nil
				}
				s1 := services.InitUserService()
				if err = s1.SetClient(gsuiteClient); err != nil {
					return nil
				}

				if err = setServiceToAction(s, a); err != nil {
					return err
				}

				if err = setServiceToAction(s1, a); err != nil {
					return err
				}
				return nil
			},
			Action: showHelpFunc,
			Subcommands: []cli.Command{
				{
					// TODO probably account command?
					Name:  "non2sv", Usage: "get employees who have not enabled 2sv",
					Action: func(context *cli.Context) error {
						return a.(*actions.LoginAction).GetNon2StepVerifiedUsers()
					},
				},
				{
					Name:  "suspicious_login", Usage: "get employees who have not been office for 30 days, but accessing",
					Action: func(c *cli.Context) error {
						activities, err := a.(*actions.LoginAction).GetAllLoginActivities(45)
						if err != nil {
							return err
						}
						return a.(*actions.LoginAction).GetIllegalLoginUsersAndIp(activities, tomlConf.GetAllIps())
					},
				},
				{
					Name:  "rare-login", Usage: "get employees who have not logged in for a while",
					Action: func(context *cli.Context) error {
						return a.(*actions.LoginAction).GetUsersWithRareLogin(14, tomlConf.Owner.DomainName)
					},
				},

			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)

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

		if strings.Contains(row[5], "@") && !strings.Contains(payload, row[5]) {
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
