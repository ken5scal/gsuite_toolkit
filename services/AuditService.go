package services

import (
	"net/http"
	"google.golang.org/api/admin/reports/v1"
)

// AuditService helps to gain insights on content management with Google Drive activity reports. Audit administrator actions.
// Generate customer and user usage reports.
// https://developers.google.com/admin-sdk/reports/v1/get-start/getting-started?authuser=1
type AuditService struct {
	*http.Client
	*admin.ActivitiesService
}

// Initialize AuditService
func InitAuditService() (*AuditService) {
	return &AuditService{}
}

// SetClient sets a client
func (s *AuditService) SetClient(client *http.Client) (error) {
	s.Client = client
	srv, err := admin.New(client)
	if err != nil{
		return err
	}
	s.ActivitiesService = srv.Activities
	return nil
}

// getAllActivities: Get All Admin Activities
// https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
func (s *AuditService) getAllActivities() {
	s.ActivitiesService.List("all", "admin")
}

// getSAMLlogin: This is just experimental
// Activities: https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
// Parameter: https://developers.google.com/admin-sdk/reports/v1/appendix/activity/login?authuser=1
// userKey: all or specific email
// applicationName: ex) login
//      choose from https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
// eventName: ex) login_failure
//      choose from https://developers.google.com/admin-sdk/reports/v1/appendix/activity/login?authuser=1#login
// filters: ex) login_type==google_password, login_failure_type<> login_failure_unknown
//      choose from https://developers.google.com/admin-sdk/reports/v1/appendix/activity/login?authuser=1#login
func (s *AuditService) getSamlLogin() {

}