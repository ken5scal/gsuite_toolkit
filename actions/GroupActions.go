package actions

import (
	"github.com/ken5scal/gsuite_toolkit/services"
	"fmt"
	"errors"
)

type GroupAction struct {
	*services.GroupService
}

func InitGroupAction() *GroupAction {
	return &GroupAction{}
}

func (a *GroupAction) SetService(s services.Service) error {
	if _, ok := s.(*services.GroupService); !ok {
		return errors.New(fmt.Sprintf("Invalid type: %T", s))
	}
	a.GroupService = s.(*services.GroupService)
	return nil
}

func (action GroupAction) RetrieveAllGroups() error {
	if g, err := action.GroupService.RetrieveAllGroups(); err != nil {
		return err
	} else {
		for _, group := range g {
			fmt.Println(group.Name + " - " + group.Email)
		}
	}
	return nil
}