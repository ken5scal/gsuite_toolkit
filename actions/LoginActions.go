package actions

import (
	"fmt"
	"errors"
	"google.golang.org/api/admin/reports/v1"
	"github.com/ken5scal/gsuite_toolkit/services"
	"github.com/ken5scal/gsuite_toolkit/utilities"
	"time"
	"github.com/asaskevich/govalidator"
)

type LoginAction struct {
	activity *services.AuditActivitiesService
	user     *services.UserService
}

func InitLoginAction() *LoginAction {
	return &LoginAction{}
}

func (action *LoginAction) SetService(s services.Service) error {
	_, ok := s.(*services.AuditActivitiesService)
	_, ok2 := s.(*services.UserService)
	if !(ok || ok2) {
		return errors.New(fmt.Sprintf("Invalid type: %T", s))
	} else if ok {
		action.activity = s.(*services.AuditActivitiesService)
	} else if ok2 {
		action.user = s.(*services.UserService)
	}
	return nil
}

// TODO Check Admin Login
func (action LoginAction) GetAllAdminUsers(domain string) error {
	// TODO Make this chan
	users, err := action.user.GetAllAdmins(domain)
	if err != nil {
		return err
	}
	hoge, err := action.user.GetAllDelegatedAdmins(domain)
	if err != nil {
		return err
	}
	users = append(users, hoge...)
	for _, user := range users {
		fmt.Println(user.PrimaryEmail)
	}
	return nil
}

func (action LoginAction) GetNon2StepVerifiedUsers() error {
	report, err := action.activity.Get2StepVerifiedStatusReport()
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
	activities, err := action.activity.GetLoginActivities(daysAgo)
	if err != nil {
		return nil, err
	}
	return activities, nil
}

func (action *LoginAction) GetUsersWithRareLogin(daysAgo int, name string) error {
	r, err := action.user.GetUsersWithRareLogin(daysAgo, name)
	if err != nil {
		return err
	}
	for _, user := range r {
		fmt.Println(user.PrimaryEmail)
	}
	return nil
}

func  (action *LoginAction)  GetIllegalLoginUsersAndIp2(officeIPs []string) error {
	// Todo this is bad
	// ToDo Make this chan
	// Wow this really needs to be Chan
	for _, ip := range officeIPs {
		if !govalidator.IsIPv4(ip) {
			return errors.New(fmt.Sprintf("%v is not in IPv4 format", ip))
		}
	}

	activities, err := action.activity.GetLoginActivities(45)
	if err != nil {
		return err
	}

	filteredActivities := getIllegalLoginUsersAndIHogep2(activities, officeIPs)
	if err != nil {
		return err
	}

	firstDayOfLastMonth := utilities.Last_Month.ModifyDate(time.Now())
	suspiciousActivitiesJudgedByGoogle, err :=  action.activity.GetSuspiciousLogins(firstDayOfLastMonth)
	if err != nil {
		return err
	}


	// ToDO Check if there is duplicates.
	suspiciousActivitiesJudgedByGoogle = append(suspiciousActivitiesJudgedByGoogle, filteredActivities...)
	//for _, activity := range filteredActivities {
	//	fmt.Println(activity.Actor.Email)
	//}
	return nil
}

// GetIllegalLoginUsersAndIp
// Main purpose is to detect employees who have not logged in from office for 30days
func getIllegalLoginUsersAndIHogep2(activities []*admin.Activity, officeIPs []string) []*admin.Activity {
	data := make(map[*admin.Activity]*LoginInformation)
	for _, activity := range activities {
		ip := activity.IpAddress
		if value, ok := data[activity]; ok {
			if !value.OfficeLogin {
				// If an user has logged in from not verified IP so far
				// then check if new IP is the one from office or not.
				value.OfficeLogin = containIP(officeIPs, ip)
			}
			value.LoginIPs = append(value.LoginIPs, ip)
		} else {
			data[activity] = &LoginInformation{
				activity.Actor.Email,
				containIP(officeIPs, ip),
				[]string{ip}}
		}
	}
	as := make([]*admin.Activity, 0, len(data))
	//for ac := range data {
	//	as = append(as, ac)
	//}
	for key, value := range data {
		if !value.OfficeLogin {
			fmt.Println(key)
			fmt.Print("     IP: ")
			fmt.Println(value.LoginIPs)
		}
	}
	return as
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