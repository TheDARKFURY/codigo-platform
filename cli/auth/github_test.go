package auth

import (
	"codigo/cli/config"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

const ConfigPath = ".config/codigo_test"

type CliGitHubSuite struct {
	suite.Suite
}

func (suite *CliGitHubSuite) TestAuthenticateIsOk() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Contains([]string{"/login/device/code", "/login/oauth/access_token"}, r.URL.Path)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				suite.NoError(err)
			}
		}(r.Body)

		body := make(map[string]string)
		err = json.NewDecoder(r.Body).Decode(&body)
		suite.NoError(err)

		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			suite.Contains(body, "client_id")
			suite.Contains(body, "scope")
			suite.EqualValues("Iv1.e87c6c02ab221540", body["client_id"])
			suite.EqualValues("read:user user:email", body["scope"])

			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			suite.Contains(body, "client_id")
			suite.Contains(body, "device_code")
			suite.Contains(body, "grant_type")
			suite.EqualValues("Iv1.e87c6c02ab221540", body["client_id"])
			suite.EqualValues("3584d83530557fdd1f46af8289938c8ef79f9dc5", body["device_code"])
			suite.EqualValues("urn:ietf:params:oauth:grant-type:device_code", body["grant_type"])

			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)
}

func (suite *CliGitHubSuite) TestAuthenticatePooling() {
	err := config.Load()
	suite.NoError(err)

	requestCounter := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=0&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			if requestCounter < 1 {
				//goland:noinspection ALL
				w.Write([]byte(`error=authorization_pending`))
			} else if requestCounter < 2 {
				//goland:noinspection ALL
				w.Write([]byte(`error=slow_down`))
			} else {
				//goland:noinspection ALL
				w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
			}

			requestCounter += 1
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)
}

func (suite *CliGitHubSuite) TestAuthenticateWhenTokenExpired() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`error=expired_token`))
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.EqualError(err, "the device code has expired. Please run `login` again")
	suite.Nil(accessToken)
}

func (suite *CliGitHubSuite) TestAuthenticateWhenAccessDenied() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`error=access_denied`))
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.EqualError(err, "login cancelled by user")
	suite.Nil(accessToken)
}

func (suite *CliGitHubSuite) TestAuthenticateWhenUnknownErrorIsReturned() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`error=unknown&message=internal server error 500`))
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.EqualError(err, "error=unknown&message=internal server error 500")
	suite.Nil(accessToken)
}

func (suite *CliGitHubSuite) TestAuthenticateWhenPoolingAuthStateTimeout() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=0&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`ok`))
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.EqualError(err, "timeout waiting for response. Please run `login` again")
	suite.Nil(accessToken)
}

func (suite *CliGitHubSuite) TestAuthenticateWhenResponseIsNot200() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		//goland:noinspection ALL
		w.Write([]byte(`internal server error 500`))
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.EqualError(err, "unhandled error: internal server error 500")
	suite.Nil(accessToken)
}

func (suite *CliGitHubSuite) TestPersistAccount() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Contains([]string{"/login/device/code", "/login/oauth/access_token", "/user"}, r.URL.Path)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				suite.NoError(err)
			}
		}(r.Body)

		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			suite.EqualValues("application/vnd.github+json", r.Header.Get("Accept"))
			suite.EqualValues("Bearer gho_16C7e42F292c6912E7710c838347Ae178B4a", r.Header.Get("Authorization"))
			suite.EqualValues("2022-11-28", r.Header.Get("X-GitHub-Api-Version"))

			resp, err := json.Marshal(map[string]string{
				"login":      "github_account",
				"created_at": "2008-01-14T04:33:35Z",
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", Account.AccessToken)
	suite.EqualValues("github_account", Account.Login)
	suite.EqualValues("2008-01-14T04:33:35Z", Account.CreatedAt)
	suite.EqualValues("not_set", Account.Name)
	suite.EqualValues("not_set", Account.Email)
	suite.EqualValues("not_set", Account.Location)
}

func (suite *CliGitHubSuite) TestPersistAccountWithAllTheData() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			resp, err := json.Marshal(map[string]string{
				"login":      "github_account",
				"created_at": "2008-01-14T04:33:35Z",
				"name":       "GitHub Account",
				"email":      "github_account@example.com",
				"location":   "Home",
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", Account.AccessToken)
	suite.EqualValues("github_account", Account.Login)
	suite.EqualValues("2008-01-14T04:33:35Z", Account.CreatedAt)
	suite.EqualValues("GitHub Account", Account.Name)
	suite.EqualValues("github_account@example.com", Account.Email)
	suite.EqualValues("Home", Account.Location)
}

func (suite *CliGitHubSuite) TestPersistAccountReturnsNoLoginProp() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			resp, err := json.Marshal(map[string]interface{}{
				"login": nil,
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.EqualError(err, "missing 'login' data from GitHub response")
}

func (suite *CliGitHubSuite) TestPersistAccountReturnsNoCreatedAtProp() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			resp, err := json.Marshal(map[string]interface{}{
				"login":      "github_account",
				"created_at": nil,
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.EqualError(err, "missing 'created_at' data from GitHub response")
}

func (suite *CliGitHubSuite) TestPersistAccountReturnsErrorMessage() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/device/code" {
			w.WriteHeader(http.StatusOK)
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			w.WriteHeader(http.StatusOK)
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			w.WriteHeader(http.StatusInternalServerError)
			resp, err := json.Marshal(map[string]string{
				"message": "simulated error",
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.EqualError(err, "couldn't get user: simulated error")
}

func (suite *CliGitHubSuite) TestPersistAccountReturnsUnknownError() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/device/code" {
			w.WriteHeader(http.StatusOK)
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			w.WriteHeader(http.StatusOK)
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			w.WriteHeader(http.StatusInternalServerError)
			//goland:noinspection ALL
			w.Write([]byte("unknown error"))
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.EqualError(err, "internal server error: unknown error")
}

func (suite *CliGitHubSuite) TestPersistAccountReturnsUnknownErrorInJson() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/device/code" {
			w.WriteHeader(http.StatusOK)
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			w.WriteHeader(http.StatusOK)
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			w.WriteHeader(http.StatusInternalServerError)
			resp, err := json.Marshal(map[string]string{
				"other": "simulated error",
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.EqualError(err, "internal server error: {\"other\":\"simulated error\"}")
}

func (suite *CliGitHubSuite) TestLogout() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			suite.EqualValues("application/vnd.github+json", r.Header.Get("Accept"))
			suite.EqualValues("Bearer gho_16C7e42F292c6912E7710c838347Ae178B4a", r.Header.Get("Authorization"))
			suite.EqualValues("2022-11-28", r.Header.Get("X-GitHub-Api-Version"))

			resp, err := json.Marshal(map[string]string{
				"login":      "github_account",
				"created_at": "2008-01-14T04:33:35Z",
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.EqualValues("gho_16C7e42F292c6912E7710c838347Ae178B4a", *accessToken)

	err = persistAccount(*accessToken)
	suite.NoError(err)

	err = LoadForLogout(ConfigPath)
	suite.NoError(err)

	err = logout()
	suite.NoError(err)
}

func (suite *CliGitHubSuite) TestLogoutWhenNotAuthenticated() {
	err := LoadForLogout(ConfigPath)
	suite.NoError(err)

	err = logout()
	suite.NoError(err)
}

func (suite *CliGitHubSuite) TestLoadingACorruptedFile() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Contains([]string{"/login/device/code", "/login/oauth/access_token", "/user"}, r.URL.Path)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				suite.NoError(err)
			}
		}(r.Body)

		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/login/device/code" {
			//goland:noinspection ALL
			w.Write([]byte(`device_code=3584d83530557fdd1f46af8289938c8ef79f9dc5&expires_in=900&interval=5&user_code=WDJB-MJHT&verification_uri=https%3A%2F%github.com%2Flogin%2Fdevice`))
		}

		if r.URL.Path == "/login/oauth/access_token" {
			//goland:noinspection ALL
			w.Write([]byte(`access_token=gho_16C7e42F292c6912E7710c838347Ae178B4a&token_type=bearer&scope=repo%2Cgist`))
		}

		if r.URL.Path == "/user" {
			suite.EqualValues("application/vnd.github+json", r.Header.Get("Accept"))
			suite.EqualValues("Bearer gho_16C7e42F292c6912E7710c838347Ae178B4a", r.Header.Get("Authorization"))
			suite.EqualValues("2022-11-28", r.Header.Get("X-GitHub-Api-Version"))

			resp, err := json.Marshal(map[string]string{
				"login":      "github_account",
				"created_at": "2008-01-14T04:33:35Z",
			})
			suite.NoError(err)

			//goland:noinspection ALL
			w.Write(resp)
		}
	}))
	defer server.Close()

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	accessToken, err := authenticate()
	suite.NoError(err)
	suite.NotNil(accessToken)

	err = persistAccount(*accessToken)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.NoError(err)

	data, err := json.Marshal(map[string]string{
		"access_token": "",
		"login":        "login",
		"created_at":   "created_at",
		"name":         "name",
		"email":        "email",
		"location":     "location",
	})
	suite.NoError(err)

	err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), data, 0600)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")

	data, err = json.Marshal(map[string]string{
		"access_token": "access_token",
		"login":        "",
		"created_at":   "created_at",
		"name":         "name",
		"email":        "email",
		"location":     "location",
	})
	suite.NoError(err)

	err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), data, 0600)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")

	data, err = json.Marshal(map[string]string{
		"access_token": "access_token",
		"login":        "login",
		"created_at":   "",
		"name":         "name",
		"email":        "email",
		"location":     "location",
	})
	suite.NoError(err)

	err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), data, 0600)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")

	data, err = json.Marshal(map[string]string{
		"access_token": "access_token",
		"login":        "login",
		"created_at":   "created_at",
		"name":         "",
		"email":        "email",
		"location":     "location",
	})
	suite.NoError(err)

	err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), data, 0600)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")

	data, err = json.Marshal(map[string]string{
		"access_token": "access_token",
		"login":        "login",
		"created_at":   "created_at",
		"name":         "name",
		"email":        "",
		"location":     "location",
	})
	suite.NoError(err)

	err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), data, 0600)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")

	data, err = json.Marshal(map[string]string{
		"access_token": "access_token",
		"login":        "login",
		"created_at":   "created_at",
		"name":         "name",
		"email":        "email",
		"location":     "",
	})
	suite.NoError(err)

	err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), data, 0600)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")

	// Invalid JSON
	err = os.WriteFile(path.Join(tokenSavedAtDir, tokenFileName), []byte("invalid json"), 0600)
	suite.NoError(err)

	err = Load(server.URL, server.URL, ConfigPath, false)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")

}

func (suite *CliGitHubSuite) TearDownTest() {
	fmt.Println("run")
	home, err := os.UserHomeDir()
	suite.NoError(err)

	err = os.RemoveAll(path.Join(home, ConfigPath))
	suite.NoError(err)

	Account = nil
}

func TestGitHubSuite(t *testing.T) {
	suite.Run(t, new(CliGitHubSuite))
}
