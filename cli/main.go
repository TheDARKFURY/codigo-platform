package main

import (
	"codigo/cli/auth"
	"codigo/cli/config"
	sentry2 "codigo/cli/sentry"
	"codigo/cli/solana"
	"codigo/cli/upgrade"
	"codigo/cli/utils"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	PRERELEASE = "true"

	cmd = &cobra.Command{
		Use:   "codigo",
		Short: "Código is an AI-Powered Code Generation Platform for blockchain",
		Long: "Código is an AI-Powered Code Generation Platform for blockchain developers and web3 teams that saves\n" +
			"development time and increases the security of the code across a variety of blockchains.\n" +
			"Complete documentation is available at https://docs.codigo.ai",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRunE: persistentPreRunE,
	}
)

func main() {
	//goland:noinspection ALL
	if PRERELEASE == "true" {
		utils.Warning("This is a pre-release version that contains the latest features. Not suitable for production.")
	}

	err := config.Load()

	if err != nil {
		utils.Error(fmt.Errorf("failed running the CLI %s", err))
		return
	}

	err = sentry2.StartSentry()

	if err != nil {
		utils.Error(fmt.Errorf("failed running the CLI %s", err))
		return
	}

	defer func() {
		err := recover()

		if err != nil {
			utils.Error(fmt.Errorf("internal server error"))
			sentry.CurrentHub().Recover(err)
			sentry.Flush(2 * time.Second)
			os.Exit(1)
		}
	}()

	solCmd, err := solana.Cmd()

	if err != nil {
		utils.Error(err)
		return
	}

	cmd.Version = config.Config.Version
	cmd.AddCommand(auth.Cmd())

	//goland:noinspection ALL
	if PRERELEASE != "true" {
		cmd.AddCommand(upgrade.Cmd())
	}

	cmd.AddCommand(solCmd)
	err = cmd.Execute()

	exitCode := 0

	if err != nil {
		utils.Error(err)
		exitCode = 1
	}

	sentry.Flush(2 * time.Second)
	os.Exit(exitCode)
}

func persistentPreRunE(cmd *cobra.Command, _ []string) error {
	if cmd.CalledAs() == "login" && cmd.Flags().Lookup("logout").Changed {
		return auth.LoadForLogout(".config/codigo")
	}

	return auth.Load("https://github.com", "https://api.github.com", ".config/codigo", true)
}
