package client

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"errors"
)

// Client to Carry out Admin job in GSuite
type ClientConfig struct {
	clientSecretFileName string
	scopes               []string
	domainName string
}

func CreateConfig() *ClientConfig {
	return &ClientConfig{}
}

func (config *ClientConfig) SetClientSecretFilename(clientSecretFileName string) *ClientConfig {
	config.clientSecretFileName = clientSecretFileName
	return config
}

func (config *ClientConfig) SetScopes(scopes []string) *ClientConfig {
	config.scopes = scopes
	return config
}

func (config *ClientConfig) setDomain(domainName string) *ClientConfig {
	config.domainName = domainName
	return config
}

// Build Generate New Client
func (config *ClientConfig) Build() (*http.Client, error) {
	b, err := ioutil.ReadFile(config.clientSecretFileName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read client secret file: %v", err))
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/admin-directory_v1-go-quickstart.json
	c, err := google.ConfigFromJSON(b, config.scopes...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}

	token := GetToken(c)
	return c.Client(context.Background(), token), nil
}

func GetToken(config *oauth2.Config) *oauth2.Token {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	token, err := tokenFromFile(cacheFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(cacheFile, token)
	}
	return token
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	_, err := fmt.Scan(&code)
	if err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("admin-directory_v1-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetAccessToken(client *http.Client) (string, error) {
	if _, ok := client.Transport.(oauth2.Transport); !ok {
		return nil, errors.New(fmt.Sprintf("Invalid type: %T", client.Transport))
	}

	token, err := client.Transport.(oauth2.Transport).Source.Token()
	if err != nil {
		return nil, err
	}
	return token.AccessToken, nil
}