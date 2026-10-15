# 학습 노트 수집

## 설치

대화형:

```bash
mise exec -- go run ./cmd/learning-wiki init
```

마법사에서 에이전트 선택 (Space 선택, Enter 진행). 미선택 시 템플릿만 설치.

비대화형:

```bash
mise exec -- go run ./cmd/learning-wiki init \
  --vault 'C:\Users\user\OneDrive\바탕 화면\workspace\LearningWiki' \
  --agents pi,codex
```

기존 Vault 연결에도 재실행 가능. 없는 Vault 생성은 `--create` 추가.

## 설치 파일

```text
<Vault>/templates/learning-note.md
~/.agents/skills/learning-capture/
├── SKILL.md
├── binding.json
└── scripts/capture.mjs
```

- 스킬·템플릿·보조 스크립트는 Go embed로 배포
- pi와 Codex 모두 `~/.agents/skills/` 탐색. 하나만 선택해도 다른 에이전트에서 보일 수 있음
- 에이전트 선택은 설치 요청·안내용이며 접근 격리나 에이전트별 기능 차이가 아님
- 둘을 선택해도 공통 스킬 한 벌만 설치
- `binding.json`은 초기 설정 시 사용한 로컬 config 파일의 절대 경로 저장. Vault 경로 자체는 config에서 조회
- 다른 PC에서는 init 재실행 필요. 공유 드라이브에 스킬의 기기별 binding을 복사하지 않기
- 기존 내용과 동일하면 유지. 다르면 충돌 오류로 중단, 자동 덮어쓰기·업데이트 없음
- 사용자 수정 템플릿도 재설치 시 충돌로 보존. 변경본 백업·비교 후 사용자가 정리
- `.agents` 등 설치 경로의 심볼릭 링크는 현재 미지원
- I/O 실패 시 이번 호출에서 생성한 리소스 파일 정리. 이미 준비한 Vault·설정 및 빈 설치 디렉터리는 남을 수 있음
- 동시 설치·악의적인 동시 파일 시스템 변경에 대한 보안 격리는 보장하지 않음

## 사용

pi:

```text
/reload
/skill:learning-capture 이번 대화의 Go 모듈과 패키지 학습 내용 저장
```

Codex: `learning-capture` 스킬로 이번 학습 내용을 저장해 달라고 요청. 발견되지 않으면 재시작 후 확인.

스킬이 완성한 Markdown을 보조 스크립트의 stdin으로 전달. 보조 스크립트는 모델을 호출하지 않으며 세션을 읽거나 요약하지 않음.

- Node.js가 에이전트의 PATH에 필요 (현재 개발 검증 버전: 22.23.3)
- `--template`으로 Vault의 노트 템플릿 조회
- UTF-8, 비어 있지 않은 노트, 최대 2 MiB
- `sessions/` 아래 타임스탬프 + UUID 파일명으로 exclusive 생성
- 성공 시 `{ "path": "저장된 절대 경로" }` 출력
- 기존 파일 덮어쓰기 없음. 재호출 시 새 노트 생성 (의미적 중복 제거 미지원)
- `sessions/` 심볼릭 링크 거부. 다른 경로로 자동 fallback 없음
- 빈 입력·잘못된 UTF-8·과대 입력·설정 누락은 실패 처리
- 직접 쓰기 방식이므로 동기화 도구에 잠시 불완전한 새 파일이 보일 수 있음. 성공 후 수동 ingest 권장
- 샌드박스가 외부 Vault 접근을 차단하면 정상 권한 승인 절차 사용

## 학습 노트 원칙

- 질문·맥락·원리·예시·혼동 해소 과정 보존
- 관련 없는 절 생략, placeholder 제거
- AI 설명과 실행·공식 문서 검증 구분
- 출처·수집 범위·누락 구간 명시, 세션 ID 추측 금지
- 원문 전체 로그 대신 정제된 근거 발췌
- 민감 정보 제외, 회사 자료 반출이 불명확하면 확인 질문
- 자동 비밀 탐지 기능 없음. 스킬의 검토 지침만으로 유출 방지가 보장되지는 않음
- `wiki/` 직접 수정 금지. 로컬 저장과 클라우드 동기화·Obsidian ingest 구분

## 검증 범위

- Go: 설치·재실행·공유 스킬 단일 설치·사용자 수정본 보존·충돌 사전 검사·심볼릭 링크 거부
- Node: 내용 보존·출력 경로·기존 파일 비덮어쓰기·템플릿 조회·입력 오류·설정 누락·sessions 심볼릭 링크 거부
- PTY: 에이전트 선택 후 최종 동의 및 설치
- 임시 HOME/Vault에서 바이너리 설치 후 다른 cwd에서 보조 스크립트 실행 검증
- 실제 pi/Codex의 스킬 발견·모델의 노트 품질·Windows Vault 연동은 사용자 환경에서 추가 검증 필요

```bash
mise exec -- go test ./...
mise exec -- node --test tests/capture.test.mjs
```

## 설치 경로 근거

- pi: 설치된 배포의 `docs/skills.md` — 사용자 `~/.agents/skills/` 및 `/reload` 지원
- Codex: https://developers.openai.com/codex/skills — USER `$HOME/.agents/skills`, 변경 미반영 시 재시작 안내
