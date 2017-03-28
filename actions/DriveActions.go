package actions

import (
	"github.com/ken5scal/gsuite_toolkit/services/drives"
	"google.golang.org/api/drive/v3"
	"fmt"
	"strconv"
	"errors"
	"github.com/ken5scal/gsuite_toolkit/services"

)

type DriveAction struct {
	*drives.Service
}

const (
	FolderMimeType = "application/vnd.google-apps.folder"
	CommandDrive         = "drive"
	SubCommandList = "list"
	SubCommandSearch = "list"
	GeneralUsage = "Audit files within Google Drive"
	ListUsage = "list all files"
	SearchUsage = "search a keyword buy specifying an argument"
)

func NewDriveAction(s services.Service) (*DriveAction, error) {
	if _, ok := s.(*drives.Service); !ok {
		return nil, errors.New(fmt.Sprintf("Invalid type: %T", s))
	}
	return &DriveAction{s.(*drives.Service)}, nil
}

func (action DriveAction) SearchFoldersWithName(title string) error {
	// 本来は'Googleフォーム'で検索したいが、検索結果が帰ってこない
	if r, err := action.GetDriveMaterialsWithTitle(title, FolderMimeType); err !=nil {
		return  err
	} else {
		for _, f := range r {
			if len(f.Parents) > 0 {
				parent, _ := action.GetParents(f.Parents[0])
				fmt.Printf(parent.Name + " > ")
			}

			fmt.Print(f.Name + "\n")
			GetPermissions(f)

			if r, err = action.GetFilesWithinDir(f.Id); err !=nil {
				return  err
			}
			if err = GetParameters(r); err != nil {
				return  err
			}
		}
	}
	return  nil
}

func (action DriveAction) SearchAllFolders() error {
	if r, err := action.GetDriveMaterialsWithTitle("*", FolderMimeType); err !=nil {
		return  err
	} else {
		for _, f := range r {
			if len(f.Parents) > 0 {
				parent, _ := action.GetParents(f.Parents[0])
				fmt.Printf(parent.Name + " > ")
			}

			fmt.Print(f.Name + "\n")
			GetPermissions(f)

			if r, err = action.GetFilesWithinDir(f.Id); err !=nil {
				return  err
			}
			if err = GetParameters(r); err != nil {
				return  err
			}
		}
	}
	return  nil
}

func GetPermissions(f *drive.File) {
	for _, p := range f.Permissions {
		fmt.Println("	" + p.Role + ": " + p.EmailAddress)
	}
}

func GetPermissions2(f *drive.File) {
	for _, p := range f.Permissions {
		fmt.Println("		" + p.Role + ": " + p.EmailAddress)
	}
}

func GetParameters(r []*drive.File) error {
	for _, report := range r {
		fmt.Println("	" + report.Name + " - " + strconv.FormatBool(report.Capabilities.CanShare))
		fmt.Println("		LastModifier: " + report.LastModifyingUser.EmailAddress)
		if len(report.Permissions) < 0 {
			return errors.New("Supposed to be no permission")
		}

		for _, o := range report.Owners {
			fmt.Println("		" + "I'm owner!" + ": " + o.EmailAddress)
		}

		GetPermissions2(report)
	}
	return nil
}