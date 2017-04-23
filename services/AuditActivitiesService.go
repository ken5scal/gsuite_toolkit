package services

import (
	"google.golang.org/api/admin/reports/v1"
	"net/http"
	"time"
	"strings"
	"google.golang.org/api/googleapi"
	"fmt"
)

/*
	Google provides Report API in purpose of auditing various Activities
	To see a list of activities to be audited, visit following:
		https://developers.google.com/admin-sdk/reports/v1/reference/activities?authuser=1
	Each activities have different event name and parameters. To understand them, visit "Reports" on bottom left of menu bar
	For example, if you want to audit Admin events on users, visit following:
	    https://developers.google.com/admin-sdk/reports/v1/reference/activity-ref-appendix-a/admin-user-events?authuser=1

	EX) Listing activities that undergoes creating new user by any user, then do following:
	  1: Visit https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
	  2: Input parameters
	        * userkey: all
	        * applicationName: admin
	        * eventName: CREATE_USER
	        * filters: USER_EMAIL==hoge@yourdomain.com
	        * startTime: 2017-04-01T00:00:00.000Z
 */

// AuditService provides following functions.
// Content management with Google Drive activity reports.
// Audit administrator actions.
// Generate customer and user usage reports.
// Details are available in a following link
// https://developers.google.com/admin-sdk/reports/
type AuditService struct {
	*admin.UserUsageReportService
	*admin.ActivitiesService
	*admin.ChannelsService
	*admin.CustomerUsageReportsService
	*http.Client
}

// Initialize AuditService
func InitAuditService() (s *AuditService) {
	return &AuditService{}
}

// SetClient creates instance of Report related Services
func (s *AuditService) SetClient(client *http.Client) (error) {
	srv, err := admin.New(client)
	if err != nil {
		return err
	}

	s.UserUsageReportService = srv.UserUsageReport
	s.ActivitiesService = srv.Activities
	s.ChannelsService = srv.Channels
	s.CustomerUsageReportsService = srv.CustomerUsageReports
	s.Client = client
	return nil
}

// getWhatever: This is just experimental so leave as it is.
// Activities: https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
// Parameter: https://developers.google.com/admin-sdk/reports/v1/appendix/activity/login?authuser=1
// userKey: all or specific email
// applicationName: ex) login
//      choose from https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
// eventName: ex) login_failure
//      choose from https://developers.google.com/admin-sdk/reports/v1/appendix/activity/login?authuser=1#login
// filters: ex) login_type==google_password, login_failure_type<> login_failure_unknown
//      choose from https://developers.google.com/admin-sdk/reports/v1/appendix/activity/login?authuser=1#login
func (s *AuditService) getWhatever() {

}

// getAllActivities: Get All Admin Activities
// https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
func (s *AuditService) getAllActivities() {
	s.ActivitiesService.List("all", "admin")
}

type RequestAuditDuration int
const (
	This_Week RequestAuditDuration = iota
	This_Month
	Last_Month
	Last_Three_Month
	Half_Year // This is the maximum duration GSuite can pull off: https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
)

// GetUserCreatedEvents lists user creation events
// Weekly, Monthly...
func (s *AuditService) GetUserCreatedEvents(d RequestAuditDuration) ([]*admin.Activity, error) {
	now := time.Now()
	switch d {
	case This_Week:
		fmt.Println("this_week")
		for now.Weekday() != time.Monday {
			now = now.AddDate(0, 0, -1)
		}
	case This_Month:
		now = now.AddDate(0, 0, -(now.Day() - 1))
	case Last_Month:
		now = now.AddDate(0, -1, -(now.Day() - 1))
	case Last_Three_Month:
		now = now.AddDate(0, -3, -(now.Day() - 1))
	case Half_Year:
		now = now.AddDate(0, -6, -(now.Day() - 1))
	}

	call := s.ActivitiesService.
		List("all", "admin").
		EventName("CREATE_USER").
		// RFC 3339 format: ex: 2010-10-28T10:26:35.000Z
		StartTime(now.Format(time.RFC3339))

	return fetchActivities(call)
}

func (s *AuditService) GetPrivilegeGrantingEvents(d RequestAuditDuration) ([]*admin.Activity, error) {
	now := time.Now()
	switch d {
	case This_Week:
		fmt.Println("this_week")
		for now.Weekday() != time.Monday {
			now = now.AddDate(0, 0, -1)
		}
	case This_Month:
		now = now.AddDate(0, 0, -(now.Day() - 1))
	case Last_Month:
		now = now.AddDate(0, -1, -(now.Day() - 1))
	case Last_Three_Month:
		now = now.AddDate(0, -3, -(now.Day() - 1))
	case Half_Year:
		now = now.AddDate(0, -6, -(now.Day() - 1))
	}

	call := s.ActivitiesService.
		List("all", "admin").
		EventName("GRANT_ADMIN_PRIVILEGE").
		StartTime(now.Format(time.RFC3339))

	return fetchActivities(call)
}

// GetUserUsage returns G Suite service activities across your account's Users
// key should be either "all" or primary id
// params should be one or combination of user report parameters
// https://developers.google.com/admin-sdk/reports/v1/guides/manage-usage-users
// Example:GetUserUsage("all", "2017-01-01", "accounts:is_2sv_enrolled,"accounts:last_name"")
func (s *AuditService) GetUserUsage(key, date, params string) (*admin.UsageReports, error) {
	return s.UserUsageReportService.
		Get(key, date).
		Parameters(params).
		Do()
}

// Get2StepVerifiedStatusReport returns reports about 2 step verification status.
// date Must be in ISO 8601 format, yyyy-mm-dd
// https://developers.google.com/admin-sdk/reports/v1/guides/manage-usage-users
// Example: Get2StepVerifiedStatusReport("2017-01-01")
func (s *AuditService) Get2StepVerifiedStatusReport() (*admin.UsageReports, error) {
	var usageReports *admin.UsageReports
	var err error
	max_retry := 10

	var timeStamp time.Time
	for i := 0; i < max_retry; i++ {
		timeStamp = time.Now().Add(-time.Duration(time.Duration(i) * time.Hour * 24))
		ts := strings.Split(timeStamp.Format(time.RFC3339), "T") // yyyy-mm-dd
		usageReports, err = s.GetUserUsage("all", ts[0], "accounts:is_2sv_enrolled")
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == http.StatusForbidden {
				return nil, err
			}
		} else if err == nil {
			break
		}
	}
	return usageReports, err
}

// GetLoginActivities reports login activities of all Users within organization
// daysAgo: number of past days you are interested from present time
// EX: GetLoginActivities(30)
func (s *AuditService) GetLoginActivities(daysAgo int) ([]*admin.Activity, error) {
	time30DaysAgo := time.Now().Add(-time.Duration(daysAgo) * time.Hour * 24)
	call := s.ActivitiesService.
		List("all", "login").
		EventName("login_success").
		StartTime(time30DaysAgo.Format(time.RFC3339))

	return fetchActivities(call)
}

func fetchActivities(call *admin.ActivitiesListCall) ([]*admin.Activity, error) {
	var activities []*admin.Activity
	for {
		if g, e := call.Do(); e != nil {
			return nil, e
		} else {
			activities = append(activities, g.Items...)
			if g.NextPageToken == "" {
				return activities, nil
			}
			call.PageToken(g.NextPageToken)
		}
	}
}