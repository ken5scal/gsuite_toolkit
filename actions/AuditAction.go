package actions

import (
	"github.com/ken5scal/gsuite_toolkit/services"
	"errors"
	"fmt"
	"github.com/ken5scal/gsuite_toolkit/utilities"
	"time"
)

// AuditAction
type AuditAction struct {
	*services.AuditActivitiesService
}

// InitGroupAction initializes Audit Action
func InitAuditAction() *AuditAction {
	return &AuditAction{}
}

// SetService sets service in Action.
func (action *AuditAction) SetService(s services.Service) error {
	if _, ok := s.(*services.AuditActivitiesService); !ok {
		return errors.New(fmt.Sprintf("Invalid type: %T", s))
	}
	action.AuditActivitiesService = s.(*services.AuditActivitiesService)
	return nil
}

// GetCreatedUserInLastMonth
func (action *AuditAction) GetCreatedUserInLastMonth() error {
	firstDayOfLastMonth := utilities.Last_Month.ModifyDate(time.Now())
	if g, err := action.AuditActivitiesService.GetUserCreatedEvents(firstDayOfLastMonth); err != nil {
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

func (action *AuditAction) GetAllGrantedPrivilegesUsersInLastMonth() error {
	firstDayOfLastMonth := utilities.Last_Month.ModifyDate(time.Now())
	activities, err := action.AuditActivitiesService.GetPrivilegeGrantedEvents(firstDayOfLastMonth)
	activities2, err2 := action.AuditActivitiesService.GetDelegatedPrivilegeGrantedEvents(firstDayOfLastMonth)
	if err != nil  {
		return err
	} else if err2 !=nil {
		return err2
	}
	activities = append(activities, activities2...)

	for _, activity := range activities {
		fmt.Println(activity.Actor)
		for _, event := range activity.Events {
			for _, parameter := range event.Parameters {
				fmt.Println(parameter.Value)
			}
		}
	}
	return nil

}