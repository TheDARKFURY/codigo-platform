package upgrade

import (
	"codigo/cli/config"
	"codigo/cli/utils"
	"fmt"
	"github.com/mouuff/go-rocket-update/pkg/provider"
	"github.com/mouuff/go-rocket-update/pkg/updater"
	"github.com/spf13/cobra"
	"runtime"
)

var (
	onlyCheck bool
	rollback  bool

	upgradeCmd = &cobra.Command{
		Use:           "upgrade",
		Short:         "Updates the Código CLI to the latest version",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: executeUpgrade,
	}
)

func executeUpgrade(_ *cobra.Command, _ []string) error {
	mUpdater := &updater.Updater{
		Provider: &provider.Github{
			RepositoryURL: config.Config.CLIUpdaterURL,
			ArchiveName:   fmt.Sprintf("bin_%s.zip", runtime.GOOS),
		},
		ExecutableName: fmt.Sprintf("codigo_%s", runtime.GOARCH),
		Version:        config.Config.Version,
	}

	if onlyCheck {
		latest, err := mUpdater.GetLatestVersion()

		if err != nil {
			return err
		}

		utils.Info(fmt.Sprintf("Latest version: %s\nInstalled version: %s", latest, config.Config.Version))
		return nil
	}

	if rollback {
		return mUpdater.Rollback()
	}

	status, err := mUpdater.Update()

	if err != nil {
		return err
	}

	switch status {
	case updater.Unknown:
		return fmt.Errorf("unknown error")
	case updater.UpToDate:
		utils.Info("You are up-to-date")
	case updater.Updated:
		utils.Info("Updated!")
	}

	return nil
}

func Cmd() *cobra.Command {
	upgradeCmd.Flags().BoolVarP(&onlyCheck, "check", "c", false, "Verifies if there is a new version available. Doesn't perform any update.")
	upgradeCmd.Flags().BoolVarP(&rollback, "rollback", "r", false, "Rollback the update. Fails if no update to rollback is available.")
	return upgradeCmd
}
