package workspace

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlwork "github.com/cppforlife/kwt/pkg/kwt/workspace"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

type SyncOptions struct {
	depsFactory   cmdcore.DepsFactory
	configFactory cmdcore.ConfigFactory
	ui            ui.UI

	WorkspaceFlags WorkspaceFlags
	SyncFlags      SyncFlags
}

func NewSyncOptions(depsFactory cmdcore.DepsFactory, configFactory cmdcore.ConfigFactory, ui ui.UI) *SyncOptions {
	return &SyncOptions{depsFactory: depsFactory, configFactory: configFactory, ui: ui}
}

func NewSyncCmd(o *SyncOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"s"},
		Short:   "Sync workspace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.WorkspaceFlags.Set(cmd, flagsFactory)
	o.SyncFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *SyncOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	restConfig, err := o.configFactory.RESTConfig()
	if err != nil {
		return err
	}

	ws := ctlwork.NewWorkspaces(o.WorkspaceFlags.NamespaceFlags.Name, coreClient)

	workspace, err := ws.Find(o.WorkspaceFlags.Name)
	if err != nil {
		return err
	}

	_ = workspace.MarkUse()

	uploadErr := UploadOperation{workspace, o.SyncFlags.Inputs, o.ui, restConfig}.Run()
	downloadErr := DownloadOperation{workspace, o.SyncFlags.Outputs, o.ui, restConfig}.Run()

	var errStrs []string

	for _, err := range []error{uploadErr, downloadErr} {
		if err != nil {
			errStrs = append(errStrs, err.Error())
		}
	}

	if len(errStrs) > 0 {
		return fmt.Errorf("- %s", strings.Join(errStrs, "\n- "))
	}

	if o.SyncFlags.Watch {
		var wg sync.WaitGroup

		for _, input := range o.SyncFlags.Inputs {
			input := input
			wg.Add(1)

			go func() {
				err := NewRemoteInput(input, UploadOperation{workspace, []ctlwork.UploadInput{input}, o.ui, restConfig}).Sync()
				fmt.Printf("error: %s\n", err)
				wg.Done()
			}()
		}

		wg.Wait()
	}

	return nil
}

type UploadOperation struct {
	Workspace  ctlwork.Workspace
	Inputs     []ctlwork.UploadInput
	UI         ui.UI
	RestConfig *rest.Config
}

func (o UploadOperation) Run() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(o.Inputs))

	for _, input := range o.Inputs {
		input := input
		wg.Add(1)

		go func() {
			o.UI.PrintLinef("[%s] Uploading input '%s'...", time.Now().Format(time.RFC3339), input.Name)

			defer func() {
				o.UI.PrintLinef("[%s] Finished uploading input '%s'...", time.Now().Format(time.RFC3339), input.Name)
			}()

			errCh <- o.Workspace.Upload(input, o.RestConfig)
			wg.Done()
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

type DownloadOperation struct {
	Workspace  ctlwork.Workspace
	Outputs    []ctlwork.DownloadOutput
	UI         ui.UI
	RestConfig *rest.Config
}

func (o DownloadOperation) Run() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(o.Outputs))

	for _, output := range o.Outputs {
		output := output
		wg.Add(1)

		go func() {
			o.UI.PrintLinef("[%s] Downloading output '%s'...", time.Now().Format(time.RFC3339), output.Name)

			defer func() {
				o.UI.PrintLinef("[%s] Finished downloading output '%s'...", time.Now().Format(time.RFC3339), output.Name)
			}()

			errCh <- o.Workspace.Download(output, o.RestConfig)
			wg.Done()
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
