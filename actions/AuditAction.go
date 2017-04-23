package actions

import (
	"github.com/ken5scal/gsuite_toolkit/services"
	"errors"
	"fmt"
)

// AuditAction
type AuditAction struct {
	*services.AuditService
}

// InitGroupAction initializes Audit Action
func InitAuditAction() *AuditAction {
	return &AuditAction{}
}

// SetService sets service in Action.
func (action *AuditAction) SetService(s services.Service) error {
	if _, ok := s.(*services.AuditService); !ok {
		return errors.New(fmt.Sprintf("Invalid type: %T", s))
	}
	action.AuditService = s.(*services.AuditService)
	return nil
}

func (action AuditAction) GetCreatedUserInLastMonth() error {
	if g, err := action.AuditService.GetUserCreatedEvents(services.Last_Month); err != nil {
		return err
	} else {
		// TODO, Wow this nest seems so unnecessary...
		for _, activity := range g {
			fmt.Println(activity.Actor)
			for _, event := range activity.Events {
				for _, parameter := range event.Parameters {
					fmt.Println(parameter.Value)
				}
			}
		}
		return nil
	}
}