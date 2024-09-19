package cmd

import (
	"bytes"
	"testing"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type cmdTestCase struct {
	name     string
	cmd      *cli.Command
	args     []string
	wantErr  bool
	wantOut  string
}

func TestCommands(t *testing.T) {
	cfg := &config.Config{}
	tests := []cmdTestCase{
		{
			name:    "Configure Command",
			cmd:     NewConfigureCmd(cfg),
			args:    []string{},
			wantErr: false,
			wantOut: "configure",
		},
		{
			name:    "Generate Command",
			cmd:     NewGenerateCmd(cfg),
			args:    []string{},
			wantErr: false,
			wantOut: "generate",
		},
		{
			name:    "Install Command",
			cmd:     NewInstallCmd(cfg),
			args:    []string{},
			wantErr: false,
			wantOut: "install",
		},
		{
			name:    "Prepare Commit Msg Command",
			cmd:     NewPrepareCommitMsgCmd(cfg),
			args:    []string{"prepare-commit-msg", "test_commit_msg_file"},
			wantErr: false,
			wantOut: "Prepare commit message hook executed successfully",
		},
		{
			name:    "Status Command",
			cmd:     NewStatusCmd(cfg),
			args:    []string{},
			wantErr: false,
			wantOut: "Status check completed",
		},
		{
			name:    "Test Command",
			cmd:     NewTestCmd(cfg),
			args:    []string{},
			wantErr: false,
			wantOut: "LLM service test completed successfully",
		},
		{
			name:    "Uninstall Command",
			cmd:     NewUninstallCmd(cfg),
			args:    []string{},
			wantErr: false,
			wantOut: "Hook uninstalled successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.App{
				Commands: []*cli.Command{tt.cmd},
			}
			buf := new(bytes.Buffer)
			app.Writer = buf
			app.ErrWriter = buf

			err := app.Run(append([]string{"app"}, tt.args...))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantOut)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tt.wantOut)
			}
		})
	}
}
