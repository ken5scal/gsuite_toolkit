package actions

import (
	"github.com/ken5scal/gsuite_toolkit/services"
	"fmt"
	"errors"
)

// GroupAction
type GroupAction struct {
	*services.GroupService
}

// InitGroupAction initializes Group
func InitGroupAction() *GroupAction {
	return &GroupAction{}
}

// SetService sets service in Action.
func (action *GroupAction) SetService(s services.Service) error {
	if _, ok := s.(*services.GroupService); !ok {
		return errors.New(fmt.Sprintf("Invalid type: %T", s))
	}
	action.GroupService = s.(*services.GroupService)
	return nil
}

// RetrieveAllGroups fetched entire group in same domain
func (action GroupAction) RetrieveAllGroups(domain string) error {
	if g, err := action.GroupService.RetrieveAllGroups(domain, ""); err != nil {
		return err
	} else {
		for _, group := range g {
			fmt.Println(group.Name + " - " + group.Email)
		}
	}
	return nil
}

// SearchGroupsByEmail searches groups where email account belongs.
func (action GroupAction) SearchGroupsByEmail(domain, email string) error {
	if g, err := action.GroupService.RetrieveAllGroups(domain, email); err != nil {
		return err
	} else {
		for _, group := range g {
			fmt.Println(group.Name + " - " + group.Email)
		}
	}
	return nil
}