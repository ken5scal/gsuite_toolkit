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
	return nil
}