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
	"github.com/twotwobread/learning-wiki/internal/hostpath"
	"github.com/twotwobread/learning-wiki/internal/obsidian"
	"github.com/twotwobread/learning-wiki/internal/resources"
	"github.com/twotwobread/learning-wiki/internal/vault"
	"golang.org/x/term"
)

var ErrCancelled = errors.New("setup cancelled")

type setupPlan struct {
	Path   string
	Create bool
	Agents []string
	Plugin bool
}
type setupPrompt interface {
	Path(context.Context) (string, error)
	Agents(context.Context) ([]string, error)
	Plugin(context.Context) (bool, error)
	Confirm(context.Context, string, string) (bool, error)
}

// planWizard collects consent without modifying the vault or configuration.
func planWizard(ctx context.Context, p setupPrompt, configPath string) (setupPlan, error) {
	return planWizardWithResolver(ctx, p, configPath, hostpath.ForCurrentHost())
}

func planWizardWithResolver(ctx context.Context, p setupPrompt, configPath string, resolver hostpath.Resolver) (setupPlan, error) {
	input, err := p.Path(ctx)
	if err != nil {
		return setupPlan{}, err
	}
	resolved, err := resolver.Resolve(ctx, input)
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
	plan.Plugin, err = p.Plugin(ctx)
	if err != nil {
		return setupPlan{}, err
	}
	approved, err := p.Confirm(ctx, "이 설정으로 초기화할까요?", fmt.Sprintf("Vault: %s\n설정: %s\n에이전트: %v\n학습 템플릿 설치 및 로컬 설정 저장\n선택 시 ~/.agents/skills/에 공통 스킬 설치 (pi·Codex 모두 탐색 가능)\n기존 파일 보존. 플러그인 설치 요청: %t (기존 설치본 유지).\n활성화·인증은 Obsidian에서 직접 수행.", resolved, configPath, plan.Agents, plan.Plugin))
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
	err := p.run(ctx, huh.NewInput().Title("Obsidian Vault 경로").Description(`로컬 절대 경로를 입력하세요. WSL에서는 Windows 드라이브 경로도 사용할 수 있습니다. 따옴표는 입력하지 마세요.`).Value(&value).Validate(func(s string) error {
		if strings.TrimSpace(s) == "" {
			return vault.ErrInvalidPath
		}
		return nil
	}))
	return value, err
}
func (p huhPrompt) Agents(ctx context.Context) ([]string, error) {
	var agents []string
	err := p.run(ctx, huh.NewMultiSelect[string]().Title("수집 스킬을 사용할 에이전트").Description("Space로 체크 · Enter로 진행. 커서 이동만으로는 선택되지 않음. 공유 경로 ~/.agents/skills/ 사용 (pi·Codex 모두 탐색). 미선택 시 템플릿만 설치.").Options(huh.NewOption("pi", "pi"), huh.NewOption("Codex", "codex")).Value(&agents))
	return agents, err
}

func (p huhPrompt) Plugin(ctx context.Context) (bool, error) {
	approved := false
	err := p.run(ctx, huh.NewConfirm().Title("Karpathy LLM Wiki 플러그인 파일을 설치할까요?").Description(fmt.Sprintf("출처: %s\n버전: %s · 라이선스: Apache-2.0\nObsidian %s 이상 필요 (앱 버전 자동 확인 없음).\n새 설치만 다운로드·SHA-256 검증. 기존 설치·인증 설정 보존.\n커뮤니티 코드는 Vault에 접근 가능. 설치만 수행하며 활성화·로그인은 직접 진행.", obsidian.Repository, obsidian.Version, obsidian.MinAppVersion)).Affirmative("설치 동의").Negative("건너뛰기").Value(&approved))
	return approved, err
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
