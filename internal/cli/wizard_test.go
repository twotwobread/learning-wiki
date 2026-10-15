package cli

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/twotwobread/learning-wiki/internal/config"
)

type scriptedPrompt struct {
	path          string
	answers       []bool
	err           error
	confirmations int
	agents        []string
}

func (p *scriptedPrompt) Path(context.Context) (string, error)     { return p.path, p.err }
func (p *scriptedPrompt) Agents(context.Context) ([]string, error) { return p.agents, p.err }
func (p *scriptedPrompt) Plugin(context.Context) (bool, error)     { return false, p.err }
func (p *scriptedPrompt) Confirm(context.Context, string, string) (bool, error) {
	p.confirmations++
	if p.err != nil {
		return false, p.err
	}
	if len(p.answers) == 0 {
		return false, errors.New("unexpected confirmation")
	}
	answer := p.answers[0]
	p.answers = p.answers[1:]
	return answer, nil
}

func TestWizardPlansWithoutWriting(t *testing.T) {
	for _, exists := range []bool{false, true} {
		t.Run(map[bool]string{false: "new", true: "existing"}[exists], func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "vault")
			if exists {
				if err := os.Mkdir(root, 0755); err != nil {
					t.Fatal(err)
				}
			}
			configPath := filepath.Join(t.TempDir(), "config.json")
			answers := []bool{true}
			if !exists {
				answers = []bool{true, true}
			}
			p := &scriptedPrompt{path: root, answers: answers, agents: []string{"pi", "codex"}}
			plan, err := planWizard(context.Background(), p, configPath)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Agents) != 2 || plan.Agents[0] != "pi" || plan.Agents[1] != "codex" {
				t.Fatalf("agent selection lost: %v", plan.Agents)
			}
			if plan.Path != root || plan.Create == exists {
				t.Fatalf("unexpected plan: %+v", plan)
			}
			if _, err := os.Stat(configPath); !errors.Is(err, fs.ErrNotExist) {
				t.Fatal("wizard wrote configuration")
			}
			if !exists {
				if _, err := os.Stat(root); !errors.Is(err, fs.ErrNotExist) {
					t.Fatal("wizard created vault before applying plan")
				}
			}
		})
	}
}

func TestWizardDeclineDoesNotWrite(t *testing.T) {
	for _, answers := range [][]bool{{false}, {true, false}} {
		root := filepath.Join(t.TempDir(), "vault")
		configPath := filepath.Join(t.TempDir(), "config.json")
		_, err := planWizard(context.Background(), &scriptedPrompt{path: root, answers: answers}, configPath)
		if !errors.Is(err, ErrCancelled) {
			t.Fatalf("expected cancellation: %v", err)
		}
		for _, path := range []string{root, configPath} {
			if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("unexpected write: %s", path)
			}
		}
	}
}

func TestWizardAbortAndConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	_, err := planWizard(context.Background(), &scriptedPrompt{err: ErrCancelled}, path)
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("lost cancellation: %v", err)
	}
	if err := config.Save(path, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	p := &scriptedPrompt{path: t.TempDir()}
	_, err = planWizard(context.Background(), p, path)
	if !errors.Is(err, config.ErrConflict) {
		t.Fatalf("expected config conflict: %v", err)
	}
	if p.confirmations != 0 {
		t.Fatal("requested approval despite existing binding conflict")
	}
}
