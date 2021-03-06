package services

import (
	"google.golang.org/api/admin/directory/v1"
	"net/http"
	"time"
	"crypto/sha1"
	"fmt"
	//"encoding/csv"
	"github.com/ken5scal/gsuite_toolkit/client"
	//"os"
	//"log"
	//"io"
	"encoding/json"
	"net/http/httputil"
	"bytes"
	"strings"

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

func (s *UserService) ConstructOuterRequest() (string, error) {
	url := "https://www.googleapis.com/batch"
	boundary := "Boundary_12345"

	token, err := client.GetAccessToken(s.Client)
	if err != nil {
		return "", err
	}

	payload := constructMultiPartMixedPayload("", boundary)
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	req.Header.Add("content-type", "multipart/mixed; boundary=" + boundary)
	req.Header.Add("authorization", "Bearer " + token)

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Dumpling Outer Request")
	fmt.Println(string(requestDump))
	fmt.Println()

	res, err := s.Client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	hoge, err := httputil.DumpResponse(res, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(hoge))

	return "", nil
}

// constructMultiPartMixedPayload constructs payload(body) as specified in rfc1341
// https://www.w3.org/Protocols/rfc1341/7_2_Multipart.html
func constructMultiPartMixedPayload(filePath, boundary string) string {
	//var reader *csv.Reader
	//var row []string
	var payload string

	header := "--" + boundary + "\nContent-Type: application/http\n\n"

	//csv_file, err := os.Open(filePath)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer csv_file.Close()
	//reader = csv.NewReader(csv_file)

	//for {
	for  i:=0; i<1 ;i++ {
		//row, err = reader.Read()
		//if err == io.EOF {
		//	return payload + "--" + boundary + "--"
		//}

		//if strings.Contains(row[5], "@") && !strings.Contains(payload, row[5]) {
		//	payload = payload + header + innerPartRequest("PUT", row[5]) + "\n\n"
		//}
		payload = payload + header + innerPartRequest(http.MethodPost, "") + "\n\n"
	}
	return payload + "--" + boundary + "--"
}

func innerPartRequest(method string, email string) (string) {
	//return "GET https://www.googleapis.com/admin/directory/v1/users/" +  email
	user := createUserObject("family3", "given", "family.given3@ken5scal01.com", "password")
	user_marshal, _ := json.Marshal(user)
	//partialResponse := "?" + "fields=users(primaryEmail,name/fullName)"
	//partialResponse := "?" + "fields=primaryEmail"

	r, _ := http.NewRequest(
		http.MethodPost,
		"https://www.googleapis.com/admin/directory/v1/users",
		bytes.NewBuffer(user_marshal))
	r.Header.Add("Content-Type", "application/json")

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Dumpling Inner Request")
	fmt.Println(string(requestDump))
	fmt.Println()
	return string(requestDump)
}

func createUserObject(familyName, givenName, email, password string) *admin.User {
	return &admin.User{
		Name: &admin.UserName{
			FamilyName: familyName,
			GivenName: givenName,
		},
		PrimaryEmail: email,
		Password: fmt.Sprintf("%x", sha1.Sum([]byte(password))),
		HashFunction: "SHA-1",
		ChangePasswordAtNextLogin: true,
	}
}

/**
POST https://www.googleapis.com/admin/directory/v1/users

{
  "name": {
    "familyName": "Family2",
    "givenName": "Given2"
  },
  "password": "5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8",
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

/* BatchRequest Example
POST /batch HTTP/1.1
Host: www.googleapis.com
Authorization: Bearer ya29.GltCBISb_LJXYXWujeQsXwKrXCjfHdFS83WV7UtQdM_MiC6mYWryO0Cu2fHga0WhjMRGz9pNS8Dy9clU27j3EaP5ywaauZnTGL3FBwaCMNgsUbLGRn5Yoxxa8i0-
Content-Type: multipart/mixed; boundary=Boundary_12345

--Boundary_12345
Content-Type: application/http

POST /admin/directory/v1/users HTTP/1.1
Host: www.googleapis.com
Content-Type: application/json

{"changePasswordAtNextLogin":true,"hashFunction":"SHA-1","name":{"familyName":"family3","givenName":"given"},"password":"5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8","primaryEmail":"family.given3@ken5scal01.com"}

--Boundary_12345--
 */

/* Batch Response
HTTP/2.0 200 OK
Alt-Svc: quic=":443"; ma=2592000; v="37,36,35"
Cache-Control: private, max-age=0
Content-Type: multipart/mixed; boundary=batch_0gBfM7LRVfc_AAWbdprRkDA
Date: Sat, 06 May 2017 12:34:53 GMT
Expires: Sat, 06 May 2017 12:34:53 GMT
Server: GSE
Vary: Origin
Vary: X-Origin
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
X-Xss-Protection: 1; mode=block

--batch_0gBfM7LRVfc_AAWbdprRkDA
Content-Type: application/http

HTTP/1.1 409 Conflict
Content-Type: application/json; charset=UTF-8
Date: Sat, 06 May 2017 12:34:53 GMT
Expires: Sat, 06 May 2017 12:34:53 GMT
Cache-Control: private, max-age=0
Content-Length: 192

{
 "error": {
  "errors": [
   {
    "domain": "global",
    "reason": "duplicate",
    "message": "Entity already exists."
   }
  ],
  "code": 409,
  "message": "Entity already exists."
 }
}

--batch_0gBfM7LRVfc_AAWbdprRkDA--
 */