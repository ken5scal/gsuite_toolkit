package main

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ken5scal/gsuite_toolkit/actions"
	"github.com/ken5scal/gsuite_toolkit/client"
	"github.com/ken5scal/gsuite_toolkit/models"
	"github.com/urfave/cli"
	"log"
	"net/http"
	"os"
	"sort"
	"github.com/ken5scal/gsuite_toolkit/services"
	"github.com/asaskevich/govalidator"
)

const (
	ClientSecretFileName = "client_secret.json"
)

type network struct {
	Name string
	Ip   []string
}

func main() {
	var tomlConf models.TomlConfig
	var service services.Service
	var action actions.Action
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
			Name: "audit", Category: "audit",
			Usage: "Audit hogehoge",
			Before: func(context *cli.Context) error {
				service = services.InitAuditService()
				if err = service.SetClient(gsuiteClient); err != nil {
					return nil
				}
				action = actions.InitAuditAction()
				return setServiceToAction(service, action)
			},
			Subcommands: []cli.Command {
				{
					Name: "user_created",
					Action: func(context *cli.Context) error {
						return action.(*actions.AuditAction).GetCreatedUserInLastMonth()
					},
				},
				{
					Name: "privileges_granted",
					Action: func(context *cli.Context) error {
						return action.(*actions.AuditAction).GetAllGrantedPrivilegesUsersInLastMonth()
					},
				},
				{
					Name:  "suspicious-login", Usage: "get employees who have not been office for 30 days, but accessing",
					Action: func(c *cli.Context) error {
						// TODO THis is Bad
						// Todo should check which service is set?
						action = actions.InitLoginAction()
						if err != setServiceToAction(service, action) {
							return err
						}
						s := services.InitUserService()
						if err = s.SetClient(gsuiteClient); err != nil {
							return nil
						}
						if err != setServiceToAction(s, action) {
							return err
						}
						return action.(*actions.LoginAction).GetIllegalLoginUsersAndIp2(tomlConf.GetAllIps())
					},
				},
				{
					Name:  "rare-login", Usage: "get employees who have not logged in for action while",
					Action: func(context *cli.Context) error {
						return action.(*actions.LoginAction).GetUsersWithRareLogin(14, tomlConf.Owner.Domain)
					},
				},
			},
		},
		{
			Name: "group", Category: "group",
			Usage: "Audit and manage groups within GSuite",
			Before: func(context *cli.Context) error {
				service = services.InitGroupService()
				if err = service.SetClient(gsuiteClient); err != nil {
					return nil
				}
				action = actions.InitGroupAction()
				return setServiceToAction(service, action)
			},
			Action: showHelpFunc,
			Subcommands: []cli.Command {
				{
					Name: "list",
					Usage: "list existing groups. By adding email address, it fetches groups where that email account belongs to.",
					Action: func(context *cli.Context) error {
						return action.(*actions.GroupAction).RetrieveAllGroups(tomlConf.Owner.Domain)
					},
				},
				{
					Name: "group",
					Usage: "get group",
					Action: func(context *cli.Context) error {
						if context.NArg() != 1 {
							return errors.New("Too few argument. Specify email.")
						} else if !govalidator.IsEmail(context.Args()[0]) {
							return errors.New("Wrong email format.")
						}
						return action.(*actions.GroupAction).GetGroup(context.Args()[0])
					},
				},
				{
					Name: "search",
					Usage: "search groups by member's email.",
					Action: func(context *cli.Context) error {
						if context.NArg() != 1 {
							return errors.New("Too few argument. Specify email.")
						} else if !govalidator.IsEmail(context.Args()[0]) {
							return errors.New("Wrong email format.")
						}
						return action.(*actions.GroupAction).SearchGroupsByEmail(tomlConf.Owner.Domain, context.Args()[0])
					},
				},
			},
		},
		{
			Name: "drive", Category: "drive",
			Usage: "Audit files within Google Drive",
			Before: func(*cli.Context) error {
				service = services.InitDriveService()
				if err = service.SetClient(gsuiteClient); err != nil {
					return nil
				}
				action = actions.InitDriveAction()
				return setServiceToAction(service, action)
			},
			Action: showHelpFunc,
			Subcommands: []cli.Command{
				{
					Name: "list", Usage: "list existing files",
					Action: func(context *cli.Context) error {
						return action.(*actions.DriveAction).SearchAllFolders()
					},
				},
				{
					Name: "search", Usage: "search action keyword buy specifying an argument",
					Action: func(context *cli.Context) error {
						if context.NArg() != 1 {
							return errors.New("Number of keyword must be exactly 1")
						}
						return action.(*actions.DriveAction).SearchFoldersByName(context.Args()[0])
					},
				},
			},
		},
		{
			Name: "user", Category: "user",
			Before: func(*cli.Context) error {
				action = actions.InitLoginAction()
				service = services.InitAuditService()
				if err = service.SetClient(gsuiteClient); err != nil {
					return nil
				}
				s1 := services.InitUserService()
				if err = s1.SetClient(gsuiteClient); err != nil {
					return nil
				}

				if err = setServiceToAction(service, action); err != nil {
					return err
				}

				if err = setServiceToAction(s1, action); err != nil {
					return err
				}
				return nil
			},
			Action: showHelpFunc,
			Subcommands: []cli.Command{
				{
					Name: "list_admin",
					Action: func(context *cli.Context) error {
						return action.(*actions.LoginAction).GetAllAdminUsers(tomlConf.Owner.Domain)
					},
				},
				{
					// TODO probably account command?
					Name:  "non2sv", Usage: "get employees who have not enabled 2sv",
					Action: func(context *cli.Context) error {
						return action.(*actions.LoginAction).GetNon2StepVerifiedUsers()
					},
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)
}