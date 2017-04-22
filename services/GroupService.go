package services

import (
	"google.golang.org/api/admin/directory/v1"
	"net/http"
)

// GroupService provides
// Details are available in a folling link
// https://developers.google.com/admin-sdk/directory/v1/guides/manage-groups
type GroupService struct {
	*admin.GroupsService
	*http.Client
}

// InitGroupService() creates a new instance
func InitGroupService() (*GroupService) {
	return &GroupService{}
}

// SetClient sets client and initialize services
func (s *GroupService) SetClient(client *http.Client) (error) {
	srv, err := admin.New(client)
	if err != nil {
		return err
	}
	s.GroupsService = srv.Groups
	s.Client = client
	return nil
}

// RepeatCallerUntilNoPageToken repeats service call until next token gets empty
func (s *GroupService) RepeatCallerUntilNoPageToken() error {
	return nil
}