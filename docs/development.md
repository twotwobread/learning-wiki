# 개발 환경

## 지원 환경 및 버전 관리

WSL Linux x64와 macOS arm64에서 개발합니다. Go와 Node.js는 mise로 관리하며 프로젝트 루트의 `mise.toml`에 패치 버전까지 고정합니다.

- Go: 1.27.2
- Node.js: 22.23.3 (기존 환경과 동일한 버전으로 시작)
- Python: 현재 프로젝트에서 사용하지 않으므로 설치·선언하지 않음

mise는 개발 환경용입니다. 배포된 npm 패키지의 최종 사용자는 Node.js/npm만 필요하며 Go나 mise를 설치할 필요가 없습니다.

## 새 개발 환경 준비

공식 설치 안내: https://mise.jdx.dev/getting-started.html

mise를 설치하고 사용하는 셸에 맞게 활성화합니다. 다음은 Bash와 zsh 예시입니다.

```bash
if [ -x "$HOME/.local/bin/mise" ]; then
    eval "$("$HOME/.local/bin/mise" activate bash)"
fi
```

```zsh
eval "$(mise activate zsh)"
```

`~/.local/bin`이 PATH에 있어야 합니다. 새 터미널을 열고 저장소에서 설정을 검토한 뒤 실행합니다.

```bash
mise trust
mise install
mise exec -- go version
mise exec -- node --version
mise exec -- npm --version
```

비대화형 셸이나 CI에서는 `.bashrc` 활성화에 의존하지 말고 `mise exec -- <명령>`을 사용합니다. 셸 연동 없이 `mise doctor`를 실행하면 비활성화 경고가 나올 수 있습니다.

전역 기본 버전은 설정하지 않습니다. 프로젝트 밖의 기존 pi용 Node 설치도 변경하지 않습니다. GOROOT는 수동 설정하지 않습니다.

## 모듈

```text
github.com/twotwobread/learning-wiki
```

CLI 명령 처리는 Cobra 사용. 의존성 버전·해시는 `go.mod`와 `go.sum`으로 관리. npm scope는 `@suyoung9137` 사용. 자체 코드 MIT 라이선스, 공개 배포 설정. 레지스트리 제공 여부는 `npm view`로 확인.

## CLI 실행·빌드

```bash
mise exec -- go run ./cmd/learning-wiki --help
mise exec -- go run ./cmd/learning-wiki --version
mise exec -- go build -o bin/learning-wiki ./cmd/learning-wiki
./bin/learning-wiki --help
```

- 인자 없음: 도움말 출력
- `--help`, `-h`: 도움말 출력
- `--version`, `-v`: 버전 출력 (기본 `dev`)
- 잘못된 명령·옵션: 오류를 stderr로 출력, 종료 코드 `2`
- 성공: 종료 코드 `0`, 마법사 취소: `130`, 그 외 실행 오류: `1`
- 터미널에서 `init` 단독 실행: Huh 기반 대화형 마법사 (경로 입력 → 필요 시 생성 동의 → 에이전트 선택 → 플러그인 설치 동의 → 최종 확인)
- 최종 확인 전 Vault·설정 쓰기 없음. 확인 기본값은 취소
- stdin 또는 stdout이 터미널이 아니면 인자 없는 `init`은 종료 코드 `2`; `--vault` 사용 안내
- `--vault` 지정 시 기존 비대화형 흐름 유지. `--create` 단독 사용·빈 `--vault`는 입력 오류
- `init --vault <절대 경로>`: 기존 로컬 Vault 연결 및 `sessions/`, `templates/` 준비
- `--create`: 없는 Vault 및 부모 디렉터리 생성 허용
- WSL에서 Windows 드라이브 절대 경로는 `wslpath -u`로 변환. WSL 절대 경로도 지원
- macOS와 일반 Linux에서는 로컬 절대 경로 사용. Windows 드라이브 입력은 거부
- 상대 경로·UNC 경로 미지원. WSL 경로가 실제 Windows 마운트인지는 강제하지 않음
- 관리 하위 디렉터리의 심볼릭 링크는 거부. 동시 외부 변경에 대한 보안 격리 보장은 아님
- 기기별 설정: Linux는 `$XDG_CONFIG_HOME/learning-wiki/config.json` (미설정 시 `~/.config/learning-wiki/config.json`), macOS는 `~/Library/Application Support/learning-wiki/config.json`
- 같은 Vault 재연결 허용. 다른 Vault로의 설정 덮어쓰기 및 손상된 설정 교체는 거부
- 학습 노트 템플릿은 기본 설치. `--agents pi,codex`로 공통 수집 스킬 설치
- pi/Codex 공유 경로 `~/.agents/skills/` 사용. 미선택 에이전트에서도 보일 수 있음
- 스킬 보조 스크립트 실행에 Node.js 필요. 상세: [학습 노트 수집](capture.md)
- `--plugin`: 고정 버전 플러그인 파일 다운로드·설치 동의. 생략 시 건너뛰기
- 기존 플러그인·인증·활성화 설정 보존. 상세: [Obsidian 연결](obsidian.md)

```bash
# 대화형 초기 설정 (터미널 필요)
mise exec -- go run ./cmd/learning-wiki init

# 기존 Windows 폴더 연결: Bash에서는 작은따옴표로 역슬래시 보존
mise exec -- go run ./cmd/learning-wiki init --vault 'D:\Notes\LearningWiki'
# 없는 폴더 생성 허용
mise exec -- go run ./cmd/learning-wiki init --vault /mnt/d/Notes/LearningWiki --create
# macOS arm64의 로컬 Vault
mise exec -- go run ./cmd/learning-wiki init --vault "$HOME/Documents/LearningWiki" --create
```

검사 실패 시 새로 생성한 빈 Vault 디렉터리는 정리. 기존 데이터는 보존. Vault 준비 이후 설정 저장에 실패하면 준비된 디렉터리는 유지하고 오류 반환; 재실행 가능. 설정 저장은 같은 디렉터리의 hard link 기반 no-replace 방식 사용. 동시 초기 설정 실행은 지원 대상으로 보장하지 않음.

빌드 버전 주입:

```bash
mise exec -- go build -ldflags '-X main.version=0.2.0' -o bin/learning-wiki ./cmd/learning-wiki
```

검증:

```bash
mise exec -- go test ./...
mise exec -- go vet ./...
mise exec -- npm test
mise exec -- npm run pack:local
mise exec -- npm run test:package
# C 컴파일러 및 CGO 활성화 필요
CGO_ENABLED=1 mise exec -- go test -race ./...
```

## 이전 WSL 환경 검증

- 공식 mise 설치 스크립트를 검토하고 체크섬 검증을 거쳐 설치.
- mise로 Go와 Node 설치 및 버전 확인 완료.
- mise 환경에서 임시 Go 프로그램 컴파일·실행 성공.
- Bash 활성화 후 `mise doctor` 정상 확인.
- 기존 수동 설치 Go와 해당 링크만 제거, pi용 Node는 보존.
- CLI 테스트·`go vet`·Linux 바이너리 빌드 및 종료 코드 검증 통과.
- 현재 환경에 C 컴파일러가 없어 race 검사는 미실행. 일반 테스트 통과와 구분.

## 추가 검증 범위

로컬 테스트는 임시 Vault·HOME에서 파일 설치와 재실행을 확인합니다. 실제 WSL/Windows 및 Mac Obsidian 앱에서의 플러그인 활성화·ingest, Codex 세션 수집 품질, 클라우드 동기화는 별도 검증 대상으로 기록합니다. npm `0.2.0` 공개 배포는 이 변경에 포함하지 않습니다.
