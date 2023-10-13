package auth

import (
	"bytes"
	"codigo/cli/config"
	"codigo/cli/sentry"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const tokenFileName = "auth_token.json"

var (
	Account         *AccountStruct
	tokenSavedAtDir string
	githubApiUrl    string
	githubUrl       string

	logoutFlag bool
	whoamiFlag bool

	login = &cobra.Command{
		Use:   "login",
		Short: "Login using your GitHub account",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: authenticateCmd,
	}
)

type AccountStruct struct {
	AccessToken string `json:"access_token"`
	Login       string `json:"login"`
	CreatedAt   string `json:"created_at"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Location    string `json:"location"`
}

func authenticateCmd(_ *cobra.Command, _ []string) error {
	if logoutFlag {
		err := logout()

		if err != nil {
			return err
		}

		fmt.Println("Logout successfully!")
	} else if whoamiFlag {
		if Account == nil {
			return fmt.Errorf("unauthorized, please authenticate by executing the command \"codigo login\"")
		}

		fmt.Printf("%s\n", Account.Login)
	} else if Account == nil {
		accessToken, err := authenticate()

		if err != nil {
			sentry.ReportGenericError(err)
			return fmt.Errorf("failed to login: %s", err)
		}

		if err := persistAccount(*accessToken); err != nil {
			return fmt.Errorf("failed to login: %s", err)
		}

		fmt.Println("Login successfully!")
	} else {
		fmt.Println("Already login...")
	}

	return nil
}

func authenticate() (*string, error) {
	body, err := json.Marshal(map[string]string{
		"client_id": config.Config.GitHubClientId,
		"scope":     "read:user user:email",
	})

	if err != nil {
		return nil, err
	}

	resp, err := http.Post(githubUrl+"/login/device/code", "application/json", bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			sentry.ReportGenericError(err)
		}
	}(resp.Body)

	body, err = io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var accessToken string

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		params, err := url.Parse("?" + string(body))

		if err != nil {
			return nil, err
		}

		deviceCode := params.Query().Get("device_code")
		userCode := params.Query().Get("user_code")
		verificationCode := params.Query().Get("verification_uri")
		interval, err := strconv.ParseInt(params.Query().Get("interval"), 10, 64)

		if err != nil {
			return nil, err
		}

		expiresIn, err := strconv.ParseInt(params.Query().Get("expires_in"), 10, 64)

		if err != nil {
			return nil, err
		}

		fmt.Printf(
			"Please visit %s\nand enter the code: %s\n",
			aurora.Blue(verificationCode).Hyperlink(verificationCode),
			userCode,
		)

		timeout := time.Now().Add(time.Second * (time.Duration)(expiresIn))

		for {
			token, err := poolAccessToken(deviceCode, interval)

			if err != nil {
				return nil, err
			}

			if token != nil {
				accessToken = *token
				break
			}

			if time.Now().After(timeout) {
				return nil, fmt.Errorf("timeout waiting for response. Please run `login` again")
			}
		}
	} else {
		return nil, fmt.Errorf("unhandled error: %s", string(body))
	}

	return &accessToken, nil
}

func poolAccessToken(deviceCode string, interval int64) (*string, error) {
	body, err := json.Marshal(map[string]string{
		"client_id":   config.Config.GitHubClientId,
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	})

	if err != nil {
		return nil, err
	}

	resp, err := http.Post(githubUrl+"/login/oauth/access_token", "application/json", bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			sentry.ReportGenericError(err)
		}
	}(resp.Body)

	body, err = io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	params, err := url.Parse("?" + string(body))

	if err != nil {
		return nil, err
	}

	if params.Query().Has("error") {
		switch params.Query().Get("error") {
		case "authorization_pending":
			fallthrough
		case "slow_down":
			time.Sleep(time.Second * (time.Duration)(interval))
		case "expired_token":
			return nil, fmt.Errorf("the device code has expired. Please run `login` again")
		case "access_denied":
			return nil, fmt.Errorf("login cancelled by user")
		default:
			return nil, fmt.Errorf(string(body))
		}
	} else if params.Query().Has("access_token") {
		t := params.Query().Get("access_token")
		return &t, nil
	}

	return nil, nil
}

func logout() error {
	err := os.Remove(path.Join(tokenSavedAtDir, tokenFileName))

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func LogoutForceAndRecuperate() error {
	file := path.Join(tokenSavedAtDir, tokenFileName)
	err := os.Remove(file)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		sentry.ReportGenericError(err)
		return fmt.Errorf("failed login out, delete the file located at \"%s\" and then execute the command \"codigo login\"", file)
	}

	return errors.New("session expired, please reauthenticate by executing the command \"codigo login\"")
}

func persistAccount(accessToken string) error {
	req, err := http.NewRequest(http.MethodGet, githubApiUrl+"/user", nil)

	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			sentry.ReportGenericError(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return fmt.Errorf("internal server error: %s", string(body))
	}

	if resp.StatusCode == 200 {
		if data["login"] == nil {
			return fmt.Errorf("missing 'login' data from GitHub response")
		}

		if data["created_at"] == nil {
			return fmt.Errorf("missing 'created_at' data from GitHub response")
		}

		name := "not_set"
		email := "not_set"
		location := "not_set"

		if data["name"] != nil {
			name = data["name"].(string)
		}

		if data["email"] != nil {
			email = data["email"].(string)
		}

		if data["location"] != nil {
			location = data["location"].(string)
		}

		acc, err := json.Marshal(&AccountStruct{
			AccessToken: accessToken,
			Login:       data["login"].(string),
			CreatedAt:   data["created_at"].(string),
			Name:        name,
			Email:       email,
			Location:    location,
		})

		if err != nil {
			return err
		}

		err = os.MkdirAll(tokenSavedAtDir, 0700)

		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), acc, 0600)

		if err != nil {
			return err
		}

		return nil
	}

	if msg, ok := data["message"]; ok {
		return fmt.Errorf("couldn't get user: %s", msg)
	}

	return fmt.Errorf("internal server error: %s", string(body))
}

func setSentryUser() {
	sentry.SetUser(Account.Login, map[string]string{
		"name":       Account.Name,
		"email":      Account.Email,
		"location":   Account.Location,
		"created_at": Account.CreatedAt,
	})
}

func LoadForLogout(tokenSavedAt string) error {
	home, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	tokenSavedAtDir = path.Join(home, tokenSavedAt)

	return nil
}

func Load(githubURL, githubApiURL, tokenSavedAt string, withSentry bool) error {
	githubUrl = githubURL
	githubApiUrl = githubApiURL
	home, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	tokenSavedAtDir = path.Join(home, tokenSavedAt)
	filePath := path.Join(tokenSavedAtDir, tokenFileName)
	_, err = os.Stat(filePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	data, err := os.ReadFile(filePath)

	if err != nil {
		sentry.ReportGenericError(err)
		return LogoutForceAndRecuperate()
	}

	err = json.Unmarshal(data, &Account)

	if err != nil {
		sentry.ReportGenericError(err)
		return LogoutForceAndRecuperate()
	}

	if len(strings.TrimSpace(Account.AccessToken)) <= 0 {
		sentry.ReportGenericError(errors.New("github \"access_token\" is empty"))
		return LogoutForceAndRecuperate()
	}

	if len(strings.TrimSpace(Account.Login)) <= 0 {
		sentry.ReportGenericError(errors.New("github \"login\" is empty"))
		return LogoutForceAndRecuperate()
	}

	if len(strings.TrimSpace(Account.CreatedAt)) <= 0 {
		sentry.ReportGenericError(errors.New("github \"created_at\" is empty"))
		return LogoutForceAndRecuperate()
	}

	if len(strings.TrimSpace(Account.Name)) <= 0 {
		sentry.ReportGenericError(errors.New("github \"name\" is empty"))
		return LogoutForceAndRecuperate()
	}

	if len(strings.TrimSpace(Account.Email)) <= 0 {
		sentry.ReportGenericError(errors.New("github \"email\" is empty"))
		return LogoutForceAndRecuperate()
	}

	if len(strings.TrimSpace(Account.Location)) <= 0 {
		sentry.ReportGenericError(errors.New("github \"location\" is empty"))
		return LogoutForceAndRecuperate()
	}

	if withSentry {
		setSentryUser()
	}

	return nil
}

func Cmd() *cobra.Command {
	login.Flags().BoolVarP(&logoutFlag, "logout", "", false, "Sign out, the user will need to execute login again.")
	login.Flags().BoolVarP(&whoamiFlag, "whoami", "", false, "Prints the authenticated username")
	return login
}
