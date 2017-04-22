package services

import "google.golang.org/api/admin/directory/v1"

// GroupService provides
// Details are available in a folling link
// https://developers.google.com/admin-sdk/directory/v1/guides/manage-groups
type GroupService struct {
	*admin.Group
	*admin.Groups
}