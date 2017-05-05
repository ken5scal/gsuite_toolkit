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
	*admin.VerificationCodesService
	*http.Client
	listCall *admin.UsersListCall
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
	s.VerificationCodesService = srv.VerificationCodes
	s.UsersService = srv.Users
	s.Client = client
	s.listCall = s.UsersService.List().OrderBy("email")
	return nil
}

// GetAllAdmins return all Admins
func (s *UserService) GetAllAdmins(domain string) ([]*admin.User, error) {
	call := s.listCall.Domain(domain).Query("isAdmin=true")
	return fetchAllUsers(call)
}

// GetAllAdmins return all Admins
func (s *UserService) GetAllDelegatedAdmins(domain string) ([]*admin.User, error) {
	call := s.listCall.Domain(domain).Query("isDelegatedAdmin=true")
	return fetchAllUsers(call)
}

// GetSuspendedEmployees retrieves users who are suspended because one of following reason:
// https://developers.google.com/admin-sdk/directory/v1/reference/users?authuser=1#resource
func (s *UserService) GetSuspendedEmployees(domain string) ([]*admin.User, error) {
	call := s.listCall.Domain(domain).Query("isSuspended=true")
	return fetchAllUsers(call)
}

// GetNon2SVEmployees retrieves users who is not using 2sv for its login,
func (s *UserService) GetNon2SVEmployees(domain string) ([]*admin.User, error) {
	call := s.listCall.Domain(domain).Query("isEnforcedIn2Sv=false isEnrolledIn2Sv=false")
	return fetchAllUsers(call)
}

// GetEmployees retrieves employees from Gsuite organization.
// By Default customer key should be "my_customer"
// max should be integer lower than 500
func (s *UserService) GetEmployees(domain string) ([]*admin.User, error) {
	call := s.listCall.Domain(domain)
	return fetchAllUsers(call)
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
	users, err := s.GetEmployees(domain)
	if err != nil {
		return nil, err
	}

	time30DaysAgo := time.Now().Add(-time.Duration(days) * time.Hour * 24)

	var goneUsers []*admin.User
	for _, user := range users {
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

// GetVerificationCodes returns verification code of user
func (s *UserService) GetVerificationCodes(email string) ([]*admin.VerificationCode, error) {
	vs,err := s.VerificationCodesService.List(email).Do()
	if err != nil {
		return nil, err
	}
	return vs.Items, nil
}

// Generate generates verification codes associated with email.
func (s *UserService) GenerateCodes(email string) error {
	return s.VerificationCodesService.Generate(email).Do()
}

// InvalidateCodes invalidates all verification codes associated with email
func (s *UserService) InvalidateCodes(email string) error {
	return s.VerificationCodesService.Invalidate(email).Do()
}

// fetchAllUsers fetches all Users
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

func requestLine(method string, email string) string {
	//return "GET https://www.googleapis.com/admin/directory/v1/users/" +  email
	return method + " " + "https://www.googleapis.com/admin/directory/v1/users/" + email + "\n" +
		"Content-Type: application/json\n\n" + body()
}

func body() string {
	return "{\n" + "\"orgUnitPath\": \"/社員・委託社員・派遣社員・アルバイト\"\n" + "}\n"
}

func createUser(familyName, givenName, emaiil, domain string) {

}

/**
POST https://www.googleapis.com/admin/directory/v1/users

{
  "name": {
    "familyName": "Family2",
    "givenName": "Given2"
  },
  "password": "ae2b657b7bcf5d2404aef5b718d96c37f974b8aa",
  "primaryEmail": "family2.given2@ken5scal01.com",
  "hashFunction": "SHA-1",
  "changePasswordAtNextLogin": true,
  "emails": [
    {
      "address": "kengoscal@gmail.com",
      "type": "other",
    },
    {
      "address": "suzuki.kengo@moneyforward.co.jp",
      "type": "other",
    },
  ]
}
 */