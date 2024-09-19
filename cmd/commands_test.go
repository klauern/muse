package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type cmdTestCase struct {
	name     string
	cmd      *cobra.Command
	args     []string
	wantErr  bool
	wantOut  string
}

func TestCommands(t *testing.T) {
	tests := []cmdTestCase{
		{
			name:    "Configure Command",
			cmd:     NewConfigureCmd(),
			args:    []string{},
			wantErr: false,
			wantOut: "Configuration completed successfully\n",
		},
		{
			name:    "Generate Command",
			cmd:     NewGenerateCmd(),
			args:    []string{},
			wantErr: false,
			wantOut: "Commit message generated successfully\n",
		},
		{
			name:    "Install Command",
			cmd:     NewInstallCmd(),
			args:    []string{},
			wantErr: false,
			wantOut: "Hook installed successfully\n",
		},
		{
			name:    "Prepare Commit Msg Command",
			cmd:     NewPrepareCommitMsgCmd(),
			args:    []string{},
			wantErr: false,
			wantOut: "Prepare commit message hook executed successfully\n",
		},
		{
			name:    "Status Command",
			cmd:     NewStatusCmd(),
			args:    []string{},
			wantErr: false,
			wantOut: "Status check completed\n",
		},
		{
			name:    "Test Command",
			cmd:     NewTestCmd(),
			args:    []string{},
			wantErr: false,
			wantOut: "LLM service test completed successfully\n",
		},
		{
			name:    "Uninstall Command",
			cmd:     NewUninstallCmd(),
			args:    []string{},
			wantErr: false,
			wantOut: "Hook uninstalled successfully\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			tt.cmd.SetOut(buf)
			tt.cmd.SetErr(buf)
			tt.cmd.SetArgs(tt.args)

			err := tt.cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Contains(t, buf.String(), tt.wantOut)
		})
	}
}
