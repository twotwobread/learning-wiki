# Learning Wiki

WSL의 pi/Codex에서 얻은 학습 내용을 Windows Obsidian Vault에 저장하기 위한 설치 패키지.

> 상태: 설계 및 초기 저장소 구성 단계. 아직 실행 가능한 CLI나 배포된 npm 패키지는 없습니다.

## 목표 흐름

1. WSL에서 npm 패키지의 설치 마법사를 실행합니다.
2. Windows 측 로컬 Vault 경로를 연결합니다.
3. pi/Codex에서 수집 스킬을 호출해 학습 노트를 저장합니다.
4. Windows Obsidian의 Karpathy LLM Wiki 플러그인으로 노트를 ingest합니다.
5. 생성된 위키를 Obsidian에서 읽고 탐색합니다.

## 첫 버전 지원 범위

- Windows x64 + WSL Linux x64.
- Windows에서 Obsidian 실행, WSL에서 pi/Codex 및 설치 도구 실행.
- WSL에서 읽고 쓸 수 있는 Windows 측 로컬 디렉터리 사용.
- Windows 경로와 WSL 경로 입력 지원. Windows 경로는 `wslpath`로 변환 후 접근 검증.
- Go 코어 + 얇은 Node.js 실행 래퍼.
- npm 플랫폼 패키지는 우선 `linux/x64`만 지원. Windows 호스트여도 실행 바이너리는 Linux용.
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

## 저장 원칙

- Vault: `sessions/`에 학습 노트, `templates/`에 템플릿, `wiki/`는 플러그인 관리 영역.
- 기기별 절대 경로는 WSL 사용자 설정에 저장하고 공유 스킬에 하드코딩하지 않습니다.
- 비밀키와 원문 세션 로그를 기본적으로 Vault에 저장하지 않습니다.
- 설치 재실행으로 기존 문서나 설정을 덮어쓰지 않습니다.
- 기존 `.obsidian` 파일은 보존하며 변경 전에 충돌 확인 및 백업을 제공합니다.
- 플러그인은 출처·라이선스·버전·배포 파일을 검토하고 사용자 동의 후 설치합니다.
- 플러그인 활성화나 보안 동의를 우회하지 않습니다.
- 파일 접근 검사 성공을 클라우드 동기화 성공으로 표현하지 않습니다.

## 개발 계획

[구현 순서 및 완료 기준](docs/implementation-plan.md)을 참고하세요.

패키지 scope, Go module 경로, 원격 저장소 URL은 아직 정하지 않았습니다.
