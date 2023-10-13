package sentry

import (
	"codigo/cli/config"
	"github.com/getsentry/sentry-go"
)

type GenErrType = string

const (
	GenErrParsing     GenErrType = "generate-parsing-failed"
	GenErrSolClientTs GenErrType = "generate-solana-client-ts-failed"
	GenErrSolProgram  GenErrType = "generate-solana-program-failed"
)

func StartSentry() error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              config.Config.SentryDsn,
		Dist:             config.Config.Version,
		Release:          config.Config.Version,
		AttachStacktrace: true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
}

func ReportInfo(msg string, ctx map[string]sentry.Context) {
	e := sentry.NewEvent()
	e.Level = sentry.LevelInfo
	e.Message = msg
	e.Contexts = ctx
	sentry.CaptureEvent(e)
}

func ReportGenerateError(eType GenErrType, err error, filename *string, cidl *[]byte) {
	e := sentry.NewEvent()
	e.Level = sentry.LevelError
	e.Type = eType

	if filename != nil && cidl != nil {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.AddAttachment(&sentry.Attachment{
				Filename:    *filename,
				ContentType: "application/yaml",
				Payload:     *cidl,
			})
		})
	}

	e.Exception = append(e.Exception, sentry.Exception{
		Value: err.Error(),
	})

	sentry.CaptureEvent(e)
}

func ReportGenericError(err error) {
	e := sentry.NewEvent()
	e.Level = sentry.LevelError
	e.Exception = append(e.Exception, sentry.Exception{
		Value: err.Error(),
	})
	sentry.CaptureEvent(e)
}

func SetUser(username string, data map[string]string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:   username,
			Data: data,
		})
	})
}
