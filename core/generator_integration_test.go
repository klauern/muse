//go:build integration

package core

import (
	"context"
	"reflect"
	"testing"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/llm"
	"github.com/klauern/muse/rag"
)

func TestNewCommitMessageGenerator(t *testing.T) {
	type args struct {
		cfg        *config.Config
		ragService rag.RAGService
	}
	tests := []struct {
		name    string
		args    args
		want    *CommitMessageGenerator
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCommitMessageGenerator(tt.args.cfg, tt.args.ragService)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCommitMessageGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCommitMessageGenerator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommitMessageGenerator_Generate(t *testing.T) {
	type fields struct {
		LLMService llm.LLMService
		RAGService rag.RAGService
	}
	type args struct {
		ctx         context.Context
		diff        string
		commitStyle string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &CommitMessageGenerator{
				LLMService: tt.fields.LLMService,
				RAGService: tt.fields.RAGService,
			}
			got, err := g.Generate(tt.args.ctx, tt.args.diff, tt.args.commitStyle)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommitMessageGenerator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CommitMessageGenerator.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}
