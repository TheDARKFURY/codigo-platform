package solana

import (
	"codigo/cli/config"
	"codigo/cli/parser"
	"codigo/cli/sentry"
	"codigo/cli/service"
	"codigo/cli/utils"
	"fmt"
	sentry2 "github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
)

const TargetAnchor string = "anchor"
const TargetNative string = "native"

var (
	programOutputPath string
	clientOutputPath  string
	disabledFileDiff  bool

	targetAnchor bool
	onlyClient   bool
	onlyProgram  bool

	cmd = &cobra.Command{
		Use:   "solana",
		Short: "Solana sub-command to generate programs and client libraries",
	}

	generate = &cobra.Command{
		Use:   "generate",
		Short: "Generate client libraries or programs",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Args: validateGenerateArgs,
		RunE: runGenerate,
	}
)

func validateGenerateArgs(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected a path to a CIDL file, received none")
	}

	p := args[0]

	if !strings.HasSuffix(strings.TrimSpace(strings.ToLower(p)), ".yaml") &&
		!strings.HasSuffix(strings.TrimSpace(strings.ToLower(p)), ".yml") {
		return fmt.Errorf("Expected a path to a CIDL file, received %s\nValid CIDL file ends with the [yaml|yml] extension", p)
	}

	return nil
}

func runGenerate(_ *cobra.Command, args []string) error {
	sentry.ReportInfo("Generate", map[string]sentry2.Context{
		"data": {
			"args":         args,
			"targetAnchor": targetAnchor,
			"onlyClient":   onlyClient,
			"onlyProgram":  onlyProgram,
		},
	})

	if !disabledFileDiff && !utils.IsGitInstalled() {
		if err := utils.AskUserToContinue(); err != nil {
			return err
		}
	} else if disabledFileDiff {
		utils.Warning("Caution: file diff is turned off. The implemented code will be lost.")
	}

	utils.Log("Parsing CIDL...")

	filename := args[0]
	idl, file, err := parser.FromFileSystem(filename, nil)

	if err != nil {
		sentry.ReportGenerateError(sentry.GenErrParsing, err, &filename, file)
		return err
	}

	extension := TargetNative

	if targetAnchor {
		extension = TargetAnchor
	}

	msg := fmt.Sprintf("%s program and client library", extension)

	if onlyProgram && onlyClient {
		utils.Info(fmt.Sprintf("Specifying --only-client and --only-program can be omitted"))
		onlyClient = true
		onlyProgram = true
	} else if onlyClient {
		msg = fmt.Sprintf("%s client library", extension)
	} else if onlyProgram {
		msg = fmt.Sprintf("%s program", extension)
	} else {
		onlyClient = true
		onlyProgram = true
	}

	utils.Log(fmt.Sprintf("Generating %s...", msg))

	// TODO: This is temporary, anchor doesn't distinguish between client and program
	if extension == TargetAnchor {
		reader, err := service.Generate(config.Config.GenServiceUrl, fmt.Sprintf("/%s/program", extension), idl)

		if err != nil {
			sentry.ReportGenerateError(sentry.GenErrSolProgram, err, &filename, file)
			return err
		}

		writer := WriteFileForAnchor

		if disabledFileDiff {
			writer = WriteFile
		}

		if err := utils.Unzip(reader, programOutputPath, writer); err != nil {
			sentry.ReportGenerateError(sentry.GenErrSolProgram, err, &filename, file)
			return err
		}

		utils.Log(fmt.Sprintf("Generated %s...", msg))
		return nil
	}

	if onlyProgram {
		reader, err := service.Generate(config.Config.GenServiceUrl, fmt.Sprintf("/%s/program", extension), idl)

		if err != nil {
			sentry.ReportGenerateError(sentry.GenErrSolProgram, err, &filename, file)
			return err
		}

		writer := WriteFileForNative

		if disabledFileDiff {
			writer = WriteFile
		}

		if err := utils.Unzip(reader, programOutputPath, writer); err != nil {
			sentry.ReportGenerateError(sentry.GenErrSolProgram, err, &filename, file)
			return err
		}

		utils.Log(fmt.Sprintf("Generated %s...", msg))
	}

	if onlyClient {
		reader, err := service.Generate(config.Config.GenServiceUrl, fmt.Sprintf("/%s/client_ts", extension), idl)

		if err != nil {
			sentry.ReportGenerateError(sentry.GenErrSolClientTs, err, &filename, file)
			return err
		}

		if err := utils.Unzip(reader, clientOutputPath, WriteFile); err != nil {
			sentry.ReportGenerateError(sentry.GenErrSolClientTs, err, &filename, file)
			return err
		}

		utils.Log(fmt.Sprintf("Generated %s...", msg))
	}

	return nil
}

func Cmd() (*cobra.Command, error) {
	cwd, err := os.Getwd()

	if err != nil {
		sentry.ReportGenericError(err)
		return nil, err
	}

	generate.Flags().StringVarP(&programOutputPath, "out-program", "", path.Join(cwd, "program"), "Output for the generated program")
	generate.Flags().StringVarP(&clientOutputPath, "out-client", "", path.Join(cwd, "program_client"), "Output for the generated client library")

	generate.Flags().BoolVarP(&targetAnchor, TargetAnchor, "a", false, "Generates the program or client using the Anchor framework")
	generate.Flags().BoolVarP(&onlyClient, "only-client", "c", false, "Generates only the TypeScript client library")
	generate.Flags().BoolVarP(&onlyProgram, "only-program", "p", false, "Generates only the program")
	generate.Flags().BoolVarP(&disabledFileDiff, "disable-diff", "", false, "Disables the diff process for the generated files. Caution: implemented code will be lost.")

	cmd.AddCommand(generate)

	return cmd, nil
}
