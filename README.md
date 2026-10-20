# Learning Wiki

pi/Codex에서 얻은 학습 내용을 로컬 Obsidian Vault에 저장하기 위한 설치 패키지.

> CLI·마법사, Vault 연결, 학습 자료 수집, 플러그인 설치 및 Linux x64/macOS arm64 npm 패키징 지원. 자체 코드 MIT 라이선스.

수집 스킬 설치·사용 방법: [학습 노트 수집](docs/capture.md).

설치·로컬 패키징·배포: [npm 패키지](docs/npm.md).

`0.2.0` 실행 (WSL Linux x64 또는 macOS arm64, Node.js 22 이상):

```bash
npx @suyoung9137/learning-wiki@0.2.0 init
```

레지스트리 제공 여부: `npm view @suyoung9137/learning-wiki version`.

레지스트리 발행 전에는 [로컬 tarball 검증](docs/npm.md)으로 같은 버전을 실행할 수 있습니다.

## 목표 흐름

1. 지원 환경에서 npm 패키지의 설치 마법사를 실행합니다.
2. 로컬 Vault 경로를 연결합니다.
3. pi/Codex에서 수집 스킬을 호출해 학습 노트를 저장합니다.
4. Obsidian의 Karpathy LLM Wiki 플러그인으로 노트를 ingest합니다.
5. 생성된 위키를 Obsidian에서 읽고 탐색합니다.

## 지원 환경

- WSL Linux x64: Windows에서 Obsidian 실행, WSL에서 pi/Codex 및 설치 도구 실행. WSL에서 접근 가능한 Windows 측 로컬 디렉터리 사용.
- macOS arm64: Mac에서 Obsidian과 pi/Codex 실행. Mac의 로컬 Vault 절대 경로 사용.
- WSL에서는 Windows 드라이브 경로를 `wslpath`로 변환. 일반 Linux와 macOS에서는 해당 환경의 절대 경로 사용.
- Go 코어 + 얇은 Node.js 실행 래퍼.
- npm 플랫폼 패키지는 `linux/x64`, `darwin/arm64`로 분리. Windows 호스트에서의 WSL 실행 바이너리는 Linux용.
- Obsidian 및 에이전트는 사용자가 미리 설치. 플러그인 활성화와 모델 인증도 사용자가 수행.
- 위키 갱신은 Obsidian 플러그인이 담당하며, 초기에는 수동 ingest.

Google Drive 등 동기화 제품은 의존성이 아닙니다. 설치 여부, 로그인, 미러링 및 동기화 상태를 검사하지 않습니다.

## 책임 경계

| 구성 요소 | 역할 |
| --- | --- |
| Go 설치 도구 | 환경·경로 검사, 설정, Vault 준비, 스킬·템플릿 설치, 플러그인 설치 지원 |
| Node.js 래퍼 | 플랫폼 바이너리 선택, 인자·표준 입출력·종료 코드 전달 |
| 수집 스킬 | 학습 맥락·예시·근거·검증 상태를 포함한 Markdown 작성 |
| Obsidian LLM Wiki | 노트 ingest, 위키 생성·통합·연결 |
| 사용자 선택 동기화 도구 | 파일 동기화 |

상시 실행 서버, 자체 위키 엔진, 과거 세션 전체 자동 수집, 자동 ingest는 첫 버전에 포함하지 않습니다.

## 원본 중심 설계

- 핵심 자산은 `sessions/`의 학습 원본 Markdown. 질문·예시·이해 과정·근거 품질에 집중.
- `wiki/` 및 `wiki/sources/`는 외부 플러그인이 가공한 결과. 정확성을 자동 보증하지 않으며 원본으로 확인.
- 플러그인의 위키 생성 품질을 직접 튜닝하거나 자체 엔진으로 대체하는 작업은 현재 범위 밖.
- 원본은 향후 다른 로컬 위키나 에이전트 질의응답에서도 재사용 가능하도록 유지.

플러그인 설치: [Obsidian 연결](docs/obsidian.md).

## 저장 원칙

- Vault: `sessions/`에 학습 노트, `templates/`에 템플릿, `wiki/`는 플러그인 관리 영역.
- 기기별 절대 경로는 해당 환경의 사용자 설정에 저장하고 공유 스킬에 하드코딩하지 않습니다.
- 비밀키와 원문 세션 로그를 기본적으로 Vault에 저장하지 않습니다.
- 설치 재실행으로 기존 문서나 설정을 덮어쓰지 않습니다.
- 기존 `.obsidian` 파일 보존. 기존 플러그인은 덮어쓰거나 업데이트하지 않음. 불완전한 설치는 수동 검토 안내.
- 플러그인은 출처·라이선스·버전·배포 파일을 검토하고 사용자 동의 후 설치합니다.
- 플러그인 활성화나 보안 동의를 우회하지 않습니다.
- 파일 접근 검사 성공을 클라우드 동기화 성공으로 표현하지 않습니다.

## 개발 계획

[구현 순서 및 완료 기준](docs/implementation-plan.md)을 참고하세요.

개발 환경은 mise로 관리합니다. [개발 환경 설정](docs/development.md)을 참고하세요.

Go module: `github.com/twotwobread/learning-wiki`. npm scope: `@suyoung9137`. 공개 패키지 3개와 private 개발용 루트 패키지로 구성. 라이선스: [MIT](LICENSE).
