package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/twotwobread/learning-wiki/internal/config"
	"github.com/twotwobread/learning-wiki/internal/resources"
	"github.com/twotwobread/learning-wiki/internal/vault"
	"golang.org/x/term"
)

var ErrCancelled = errors.New("setup cancelled")

type setupPlan struct {
	Path   string
	Create bool
	Agents []string
}
type setupPrompt interface {
	Path(context.Context) (string, error)
	Agents(context.Context) ([]string, error)
	Confirm(context.Context, string, string) (bool, error)
}

// planWizard collects consent without modifying the vault or configuration.
func planWizard(ctx context.Context, p setupPrompt, configPath string) (setupPlan, error) {
	input, err := p.Path(ctx)
	if err != nil {
		return setupPlan{}, err
	}
	resolved, err := vault.Resolve(ctx, input, nil)
	if err != nil {
		return setupPlan{}, err
	}
	if err = config.Check(configPath, resolved); err != nil {
		return setupPlan{}, err
	}
	plan := setupPlan{Path: resolved}
	info, err := os.Stat(resolved)
	if errors.Is(err, fs.ErrNotExist) {
		approved, err := p.Confirm(ctx, "Vault 폴더를 생성할까요?", resolved)
		if err != nil {
			return setupPlan{}, err
		}
		if !approved {
			return setupPlan{}, ErrCancelled
		}
		plan.Create = true
	} else if err != nil {
		return setupPlan{}, err
	} else if !info.IsDir() {
		return setupPlan{}, fmt.Errorf("%w: %s", vault.ErrNotDirectory, resolved)
	}
	plan.Agents, err = p.Agents(ctx)
	if err != nil {
		return setupPlan{}, err
	}
	if err = resources.ValidateAgents(plan.Agents); err != nil {
		return setupPlan{}, err
	}
	approved, err := p.Confirm(ctx, "이 설정으로 초기화할까요?", fmt.Sprintf("Vault: %s\n설정: %s\n에이전트: %v\n학습 템플릿 설치 및 로컬 설정 저장\n선택 시 ~/.agents/skills/에 공통 스킬 설치 (pi·Codex 모두 탐색 가능)\n기존 파일 보존. Obsidian 플러그인 설치는 아직 미지원.", resolved, configPath, plan.Agents))
	if err != nil {
		return setupPlan{}, err
	}
	if !approved {
		return setupPlan{}, ErrCancelled
	}
	return plan, nil
}

type huhPrompt struct {
	in  io.Reader
	out io.Writer
}

func (p huhPrompt) run(ctx context.Context, field huh.Field) error {
	err := huh.NewForm(huh.NewGroup(field)).WithInput(p.in).WithOutput(p.out).RunWithContext(ctx)
	if errors.Is(err, huh.ErrUserAborted) || errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %w", ErrCancelled, err)
	}
	return err
}
func (p huhPrompt) Path(ctx context.Context) (string, error) {
	var value string
	err := p.run(ctx, huh.NewInput().Title("Obsidian Vault 경로").Description(`Windows 절대 경로 또는 /mnt/... 경로 입력. 따옴표는 입력하지 마세요.`).Value(&value).Validate(func(s string) error {
		if strings.TrimSpace(s) == "" {
			return vault.ErrInvalidPath
		}
		return nil
	}))
	return value, err
}
func (p huhPrompt) Agents(ctx context.Context) ([]string, error) {
	var agents []string
	err := p.run(ctx, huh.NewMultiSelect[string]().Title("수집 스킬을 사용할 에이전트").Description("공유 경로 ~/.agents/skills/ 사용. 하나만 선택해도 다른 에이전트에서 보일 수 있음. 미선택 시 템플릿만 설치.").Options(huh.NewOption("pi", "pi"), huh.NewOption("Codex", "codex")).Value(&agents))
	return agents, err
}

func (p huhPrompt) Confirm(ctx context.Context, title, description string) (bool, error) {
	approved := false
	err := p.run(ctx, huh.NewConfirm().Title(title).Description(description).Affirmative("진행").Negative("취소").Value(&approved))
	return approved, err
}

func isTerminal(stream any) bool {
	f, ok := stream.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(int(f.Fd()))
}
