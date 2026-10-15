# 개발 환경

## 지원 환경 및 버전 관리

WSL Linux x64에서 개발합니다. Go와 Node.js는 mise로 관리하며 프로젝트 루트의 `mise.toml`에 패치 버전까지 고정합니다.

- Go: 1.27.2
- Node.js: 22.23.3 (기존 환경과 동일한 버전으로 시작)
- Python: 현재 프로젝트에서 사용하지 않으므로 설치·선언하지 않음

mise는 개발 환경용입니다. 배포된 npm 패키지의 최종 사용자는 Node.js/npm만 필요하며 Go나 mise를 설치할 필요가 없습니다.

## 새 개발 환경 준비

공식 설치 안내: https://mise.jdx.dev/getting-started.html

mise를 설치하고 Bash 설정에 다음을 추가합니다. 기본 사용자 설치 위치 기준입니다.

```bash
if [ -x "$HOME/.local/bin/mise" ]; then
    eval "$("$HOME/.local/bin/mise" activate bash)"
fi
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

CLI 명령 처리는 Cobra 사용. 의존성 버전·해시는 `go.mod`와 `go.sum`으로 관리. npm scope는 미정.

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
- 성공: 종료 코드 `0`, 그 외 실행 오류: `1`
- `init` 등 설치 명령은 아직 미구현

빌드 버전 주입:

```bash
mise exec -- go build -ldflags '-X main.version=0.1.0' -o bin/learning-wiki ./cmd/learning-wiki
```

검증:

```bash
mise exec -- go test ./...
mise exec -- go vet ./...
# C 컴파일러 및 CGO 활성화 필요
CGO_ENABLED=1 mise exec -- go test -race ./...
```

## 현재 환경 검증

- 공식 mise 설치 스크립트를 검토하고 체크섬 검증을 거쳐 설치.
- mise로 Go와 Node 설치 및 버전 확인 완료.
- mise 환경에서 임시 Go 프로그램 컴파일·실행 성공.
- Bash 활성화 후 `mise doctor` 정상 확인.
- 기존 수동 설치 Go와 해당 링크만 제거, pi용 Node는 보존.
- CLI 테스트·`go vet`·Linux 바이너리 빌드 및 종료 코드 검증 통과.
- 현재 환경에 C 컴파일러가 없어 race 검사는 미실행. 일반 테스트 통과와 구분.

## 다음 작업

환경 감지, 경로 변환 및 접근 검사를 외부 명령과 파일 시스템으로부터 테스트 가능한 형태로 분리하여 구현합니다. 이후 `init` 명령과 설치 마법사를 연결합니다.
