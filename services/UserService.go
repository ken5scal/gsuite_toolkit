package services

import (
	"google.golang.org/api/admin/directory/v1"
	"net/http"
	"time"
)

// UserService provides User related administration Task
// Details are available in a following link
// https://developers.google.com/admin-sdk/directory/v1/guides/manage-users
type UserService struct {
	*admin.UsersService
	*http.Client
}

// Initialize UserService
func InitUserService() (s *UserService) {
	return &UserService{}
}

// SetClient creates instance of User related Services
func (s *UserService) SetClient(client *http.Client) (error) {
	srv, err := admin.New(client)
	if err != nil {
		return err
	}
	s.UsersService = srv.Users
	s.Client = client
	return nil
}

// GetAllAdmins return all Admins
func (s *UserService) GetAllAdmins(domain string) ([]*admin.User, error) {
	call := s.UsersService.
		List().Domain(domain).OrderBy("email").
		Query("isAdmin=true")
	return fetchAllUsers(call)
}

// GetAllAdmins return all Admins
func (s *UserService) GetAllDelegatedAdmins(domain string) ([]*admin.User, error) {
	call := s.UsersService.
		List().Domain(domain).OrderBy("email").
		Query("isDelegatedAdmin=true")
	return fetchAllUsers(call)
}

func (s *UserService) GetSuspendedEmployees(domain string) ([]*admin.User, error) {
	call := s.UsersService.
		List().Domain(domain).OrderBy("email").
		Query("isSuspended=true")
	return fetchAllUsers(call)
}

// GetEmployees retrieves employees from Gsuite organization.
// By Default customer key should be "my_customer"
// max should be integer lower than 500
// ToDo This should return []*admin.User instead of *admin.Users
func (s *UserService) GetEmployees(domain string) ([]*admin.User, error) {
	call := s.UsersService.
		List().Domain(domain).OrderBy("email")
	return fetchAllUsers(call)
}

// GetAllUsersInDomain retrieves all users in domain.
// GET https://www.googleapis.com/admin/directory/v1/users?domain=example.com&maxResults=2
// Example: GetAllUsersInDomain("hoge.co.jp", "[email, familyname, givenname]", 500)
func (s *UserService) GetAllUsersInDomain(domain string, max int64) (*admin.Users, error) {
	return s.UsersService.
		List().
		Domain(domain).
		OrderBy("email").
		MaxResults(max).
		Do()
}

// GetUser retrieves a user based on either email or userID
// GET https://www.googleapis.com/admin/directory/v1/users/userKey
// Example: GetUser("abc@abc.co.jp")
func (s *UserService) GetUser(key string) (*admin.User, error) {
	return s.UsersService.Get(key).ViewType("domain_public").Do()
}

// ChangeOrgUnit changes user's OrgUnit.
// PUT https://www.googleapis.com/admin/directory/v1/users/{email/userID}
// Example: ChangeOrgUnit(user, "社員・委託社員・派遣社員・アルバイト")
func (s *UserService) ChangeOrgUnit(user *admin.User, unit string) (*admin.User, error) {
	user.OrgUnitPath = "/" + unit
	return s.UsersService.Update(user.PrimaryEmail, user).Do()
}

// GetUsersWithRareLogin detects who has not logged in recently.
func (s *UserService) GetUsersWithRareLogin(days int, domain string) ([]*admin.User, error) {
	users, err := s.GetAllUsersInDomain(domain, 500)
	if err != nil {
		return nil, err
	}

	time30DaysAgo := time.Now().Add(-time.Duration(days) * time.Hour * 24)

	var goneUsers []*admin.User
	for _, user := range users.Users {
		lastLogin, err := time.Parse("2006-01-02T15:04:05.000Z", user.LastLoginTime)
		if err != nil {
			return nil, err
		}
		if time30DaysAgo.After(lastLogin) {
			goneUsers = append(goneUsers, user)
		}
	}

	return goneUsers, nil
}

func fetchAllUsers(call *admin.UsersListCall) ([]*admin.User, error) {
	var users []*admin.User
	for {
		if g, e := call.Do(); e != nil {
			return nil, e
		} else {
			users = append(users, g.Users...)
			if g.NextPageToken == "" {
				return users, nil
			}
			call.PageToken(g.NextPageToken)
		}
	}
}