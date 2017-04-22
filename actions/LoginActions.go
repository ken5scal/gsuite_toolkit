package actions

import (
	"fmt"
	"errors"
	"google.golang.org/api/admin/reports/v1"
	"github.com/ken5scal/gsuite_toolkit/services"
)

type LoginAction struct {
	report *services.ReportService
	user *services.UserService
}

func InitLoginAction() *LoginAction {
	return &LoginAction{}
}

func (action *LoginAction) SetService(s services.Service) error {
	_, ok := s.(*services.ReportService)
	_, ok2 := s.(*services.UserService)
	if !(ok || ok2) {
		return errors.New(fmt.Sprintf("Invalid type: %T", s))
	} else if ok {
		action.report = s.(*services.ReportService)
	} else if ok2 {
		action.user = s.(*services.UserService)
	}
	return nil
}

// TODO Check Admin Login

func (action LoginAction) GetNon2StepVerifiedUsers() error {
	report, err := action.report.Get2StepVerifiedStatusReport()
	if err != nil {
		return err
	}

	if len(report.UsageReports) == 0 {
		return errors.New("No Report Available")
	}

	var paramIndex int
	fmt.Println("Latest Report: " + report.UsageReports[0].Date)
	for i, param := range report.UsageReports[0].Parameters {
		// https://developers.google.com/admin-sdk/reports/v1/guides/manage-usage-users
		// Parameters: https://developers.google.com/admin-sdk/reports/v1/reference/usage-ref-appendix-a/users-accounts
		if param.Name == "accounts:is_2sv_enrolled" {
			paramIndex = i
			break
		}
	}

	for _, r := range report.UsageReports {
		if !r.Parameters[paramIndex].BoolValue {
			fmt.Println(r.Entity.UserEmail)
		}
	}

	return nil
}

func (action LoginAction) GetAllLoginActivities(daysAgo int) ([]*admin.Activity, error) {
	activities, err := action.report.GetLoginActivities(daysAgo)
	if err != nil {
		return nil, err
	}
	return activities, nil
}

func (action LoginAction) GetUsersWithRareLogin(daysAgo int, name string) error {
	r, err := action.user.GetUsersWithRareLogin(daysAgo, name)
	if err != nil {
		return err
	}
	for _, user := range r {
		fmt.Println(user.PrimaryEmail)
	}
	return nil
}

// GetIllegalLoginUsersAndIp
// Main purpose is to detect employees who have not logged in from office for 30days
func  (action LoginAction)  GetIllegalLoginUsersAndIp(activities []*admin.Activity, officeIPs []string) error {
	data := make(map[string]*LoginInformation)
	for _, activity := range activities {
		email := activity.Actor.Email
		ip := activity.IpAddress

		if value, ok := data[email]; ok {
			if !value.OfficeLogin {
				// If an user has logged in from not verified IP so far
				// then check if new IP is the one from office or not.
				value.OfficeLogin = containIP(officeIPs, ip)
			}
			value.LoginIPs = append(value.LoginIPs, ip)
		} else {
			data[email] = &LoginInformation{
				email,
				containIP(officeIPs, ip),
				[]string{ip}}
		}
	}

	for key, value := range data {
		if !value.OfficeLogin {
			fmt.Println(key)
			fmt.Print("     IP: ")
			fmt.Println(value.LoginIPs)
		}
	}
	return nil
}
type LoginInformation struct {
	Email       string
	OfficeLogin bool
	LoginIPs    []string
}

func containIP(ips []string, ip string) bool {
	set := make(map[string]struct{}, len(ips))
	for _, s := range ips {
		set[s] = struct{}{}
	}

	_, ok := set[ip]
	return ok
}