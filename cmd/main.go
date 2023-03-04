package cmd

import (
	"adrianolaselva.github.io/csvql/cmd/csvqlctl"
	"fmt"
	"github.com/spf13/cobra"
	"syscall"
)

const (
	commandBase = "csvql"
	bannerPrint = `csvql cli tools`
)

type CliBase interface {
	Execute() error
}

type cliBase struct {
	rootCmd *cobra.Command
}

func New() CliBase {
	var release = "latest"
	if value, ok := syscall.Getenv("VERSION"); ok {
		release = value
	}

	cmd := &cobra.Command{
		Use:     commandBase,
		Version: release,
		Long:    bannerPrint,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    true,
		},
	}

	return &cliBase{rootCmd: cmd}
}

func (c *cliBase) Execute() error {
	csvQlCtl, err := csvqlctl.New().Command()
	if err != nil {
		return err
	}

	c.rootCmd.AddCommand(csvQlCtl)

	if err := c.rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute command %s", err)
	}

	return nil
}
