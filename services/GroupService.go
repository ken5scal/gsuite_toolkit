package services

import (
	"google.golang.org/api/admin/directory/v1"
	"net/http"
	"fmt"
)

// Group Descriptor
// https://developers.google.com/admin-sdk/directory/v1/reference/groups?authuser=1#resource

// GroupService provides
// Details are available in a folling link
// https://developers.google.com/admin-sdk/directory/v1/guides/manage-groups
type GroupService struct {
	*admin.GroupsService
	*http.Client
	*admin.GroupsListCall
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

// GetGroup Retrieves a group's properties
// https://developers.google.com/admin-sdk/directory/v1/reference/groups/get?authuser=1
func (s *GroupService) GetGroup(groupEmail string) (*admin.Group, error) {
	return s.GroupsService.Get(groupEmail).Do()
}

// RetrieveAllgroups Rrtrieves a paginated list of groups in a domain.
// https://developers.google.com/admin-sdk/directory/v1/reference/groups/list?authuser=1
func (s *GroupService) RetrieveAllGroups(domain, email string) ([]*admin.Group, error) {
	call := s.GroupsService.List().
		Domain(domain).
		UserKey(email)
	var groups []*admin.Group
	for {
		g, e := call.Do()
		if e != nil {
			return nil, e
		}
		groups = append(groups, g.Groups...)
		if g.NextPageToken == "" {
			return groups, nil
		}
		call.PageToken(g.NextPageToken)
	}
}