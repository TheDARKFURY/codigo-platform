package service

import (
	"archive/zip"
	"bytes"
	"codigo/cli/auth"
	"codigo/cli/parser"
	"codigo/cli/sentry"
	"codigo/cli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go/v4"
	"io"
	"net/http"
)

func Generate(serverUrl, path string, idl *parser.IDL) (*zip.Reader, error) {
	if auth.Account == nil {
		return nil, fmt.Errorf("unauthorized, please authenticate by executing the command \"codigo login\"")
	}

	body, err := json.Marshal(idl)

	if err != nil {
		return nil, err
	}

	reader, err := retry.DoWithData(
		func() (*zip.Reader, error) {
			req, err := http.NewRequest(http.MethodPost, serverUrl+"/public_beta/solana"+path, bytes.NewBuffer(body))

			if err != nil {
				return nil, err
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth.Account.AccessToken))

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				return nil, err
			}

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					sentry.ReportGenericError(err)
				}
			}(resp.Body)

			if resp.StatusCode != 200 {
				respBody, err := io.ReadAll(resp.Body)

				if err != nil {
					return nil, err
				}

				var response Response

				err = json.Unmarshal(respBody, &response)

				if err != nil {
					return nil, &ServerError{Message: string(respBody)}
				}

				switch resp.StatusCode {
				case http.StatusBadRequest:
					return nil, &BadRequest{Message: response.Message}
				case http.StatusUnauthorized:
					return nil, &UnauthorizedRequest{Message: response.Message}
				case http.StatusInternalServerError:
					return nil, &ServerError{Message: response.Message}
				default:
					return nil, &UnknownError{Message: response.Message}
				}
			}

			buffer := new(bytes.Buffer)

			_, err = io.Copy(buffer, resp.Body)

			if err != nil {
				return nil, err
			}

			reader := bytes.NewReader(buffer.Bytes())

			data, err := zip.NewReader(reader, resp.ContentLength)

			if err != nil {
				return nil, err
			}

			return data, nil
		},
		retry.OnRetry(func(n uint, err error) {
			utils.Log(fmt.Sprintf("Request failed, retrying... %s", err))
		}),
		retry.Attempts(3),
		retry.RetryIf(func(err error) bool {
			var serverError *ServerError
			return errors.As(err, &serverError)
		}),
		retry.LastErrorOnly(true),
		retry.WrapContextErrorWithLastError(false),
	)

	if err != nil {
		var unauthorized *UnauthorizedRequest

		if errors.As(err, &unauthorized) {
			return nil, auth.LogoutForceAndRecuperate()
		}

		return nil, err
	}

	return reader, nil
}
