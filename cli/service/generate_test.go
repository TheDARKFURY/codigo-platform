package service

import (
	"archive/zip"
	"bytes"
	"codigo/cli/auth"
	"codigo/cli/config"
	"codigo/cli/parser"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var account = &auth.AccountStruct{
	AccessToken: "gho_16C7e42F292c6912E7710c838347Ae178B4a",
	Login:       "github_account",
	CreatedAt:   "2008-01-14T04:33:35Z",
	Email:       "not_set",
	Location:    "not_set",
	Name:        "not_set",
}

type CliGenerateSuite struct {
	suite.Suite
}

func (suite *CliGenerateSuite) TestGenerateIsOk() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.EqualValues("/public_beta/solana/native/program", r.URL.Path)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				suite.NoError(err)
			}
		}(r.Body)

		var idl parser.IDL
		err = json.NewDecoder(r.Body).Decode(&idl)
		suite.NoError(err)

		buffer := new(bytes.Buffer)
		archive := zip.NewWriter(buffer)

		err = archive.Close()
		suite.NoError(err)

		//goland:noinspection ALL
		w.Write(buffer.Bytes())
	}))
	defer server.Close()

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate(server.URL, "/native/program", idl)
	suite.NoError(err)
	suite.NotNil(archive)
	suite.EqualValues(0, len(archive.File))
}

func (suite *CliGenerateSuite) TestGenerateCreateZipReaderReturnsErr() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Length", "-1")
		//goland:noinspection ALL
		w.Write([]byte{})
	}))
	defer server.Close()

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate(server.URL, "/native/program", idl)
	suite.EqualError(err, "zip: size cannot be negative")
	suite.Nil(archive)
}

func (suite *CliGenerateSuite) TestGenerateWhenNotAuthenticated() {
	auth.Account = nil

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate("", "/native/program", idl)
	suite.EqualError(err, "unauthorized, please authenticate by executing the command \"codigo login\"")
	suite.Nil(archive)
}

func (suite *CliGenerateSuite) TestGenerateWhenServiceReturns400() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp, err := json.Marshal(map[string]string{
			"message": "simulated bad request",
		})
		suite.NoError(err)

		//goland:noinspection ALL
		w.Write(resp)
	}))
	defer server.Close()

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate(server.URL, "/native/program", idl)
	suite.EqualError(err, "simulated bad request")
	suite.Nil(archive)
}

func (suite *CliGenerateSuite) TestGenerateWhenServiceReturns401() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp, err := json.Marshal(map[string]string{
			"message": "simulated unauthorized",
		})
		suite.NoError(err)

		//goland:noinspection ALL
		w.Write(resp)
	}))
	defer server.Close()

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate(server.URL, "/native/program", idl)
	suite.EqualError(err, "session expired, please reauthenticate by executing the command \"codigo login\"")
	suite.Nil(archive)
}

func (suite *CliGenerateSuite) TestGenerateWhenServiceReturns500() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		resp, err := json.Marshal(map[string]string{
			"message": "simulated server error",
		})
		suite.NoError(err)

		//goland:noinspection ALL
		w.Write(resp)
	}))
	defer server.Close()

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate(server.URL, "/native/program", idl)
	suite.EqualError(err, "simulated server error")
	suite.Nil(archive)
}

func (suite *CliGenerateSuite) TestGenerateWhenServiceReturnsUnhandledError() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		resp, err := json.Marshal(map[string]string{
			"message": "simulated bad gateway",
		})
		suite.NoError(err)

		//goland:noinspection ALL
		w.Write(resp)
	}))
	defer server.Close()

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate(server.URL, "/native/program", idl)
	suite.EqualError(err, "simulated bad gateway")
	suite.Nil(archive)
}

func (suite *CliGenerateSuite) TestGenerateWhenServiceReturnsNonJson() {
	err := config.Load()
	suite.NoError(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)

		//goland:noinspection ALL
		w.Write([]byte("simulated non json"))
	}))
	defer server.Close()

	idl, _, mErr := parser.FromFileSystem("../../templates/validate_accounts.yaml", nil)
	suite.Nil(mErr)

	archive, err := Generate(server.URL, "/native/program", idl)
	suite.EqualError(err, "simulated non json")
	suite.Nil(archive)
}

func (suite *CliGenerateSuite) SetupTest() {
	auth.Account = account
}

func TestCliGenerateSuite(t *testing.T) {
	suite.Run(t, new(CliGenerateSuite))
}
