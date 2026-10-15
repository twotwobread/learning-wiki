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

현재 외부 Go 의존성이 없어 `go.sum`은 없습니다. npm scope는 아직 정하지 않았습니다.

## 현재 환경 검증

- 공식 mise 설치 스크립트를 검토하고 체크섬 검증을 거쳐 설치.
- mise로 Go와 Node 설치 및 버전 확인 완료.
- mise 환경에서 임시 Go 프로그램 컴파일·실행 성공.
- Bash 활성화 후 `mise doctor` 정상 확인.
- 기존 수동 설치 Go와 해당 링크만 제거, pi용 Node는 보존.
- 애플리케이션 소스와 테스트는 아직 없으므로 프로젝트 빌드·테스트는 다음 단계에서 추가.

## 다음 작업

환경 감지, 경로 변환 및 접근 검사를 외부 명령과 파일 시스템으로부터 테스트 가능한 형태로 분리하여 구현합니다. 이후 `init` 명령과 설치 마법사를 연결합니다.
