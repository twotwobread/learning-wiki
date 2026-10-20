# Obsidian 플러그인 설치 지원

## 원칙

- 관리 대상: 외부 플러그인의 배포 파일 설치만 지원
- 관리하지 않는 대상: 플러그인 생성 엔진·프롬프트·위키 품질 튜닝, 자동 ingest, 인증·활성화 설정
- 원본 자산: `sessions/`의 학습 자료. `wiki/sources/`도 가공 결과이며 원본이 아님

## 사용

Windows 또는 macOS에서 Obsidian을 먼저 설치. 기본 설정 폴더 `.obsidian` 사용을 전제로 하며 커스텀 설정 폴더는 미지원. WSL에서는 Windows Obsidian과 같은 Vault에 접근해야 합니다.

마법사:

```bash
mise exec -- go run ./cmd/learning-wiki init
```

플러그인 설치 질문에서 출처·버전·라이선스·권한 안내 확인 후 선택. 기본값은 건너뛰기. 에이전트 선택과는 별개.

명시적 설치 동의:

```bash
mise exec -- go run ./cmd/learning-wiki init \
  --vault 'D:\Notes\LearningWiki' \
  --agents pi,codex \
  --plugin
```

macOS에서는 `--vault "$HOME/Documents/LearningWiki"`처럼 로컬 절대 경로를 사용합니다.

`--plugin` 자체가 다운로드·파일 설치 동의. 생략 시 플러그인 관련 네트워크 요청·파일 변경 없음. 없는 Vault 생성은 `--create` 추가.

## 고정된 배포본

- 프로젝트: https://github.com/GD4AI/obsidian-llm-wiki
- 버전: `1.28.0`
- ID: `karpathywiki`
- 최소 Obsidian 버전: `1.11.4`
- 라이선스: Apache-2.0; 번들된 의존성의 별도 고지는 THIRD-PARTY-NOTICES.md 참조
- 설치 위치: `<Vault>/.obsidian/plugins/karpathywiki/`
- 파일: `manifest.json`, `main.js`, `styles.css`, `LICENSE`, `NOTICE`, `THIRD-PARTY-NOTICES.md`

코드는 수정하지 않고 GitHub 공식 릴리스/태그에서 다운로드. 런타임 파일 SHA-256은 GitHub 릴리스 메타데이터와 대조하고 코드에 고정; 라이선스 문서는 동일 태그에서 조회한 해시 고정. 해시 변경 시 자동 수용 금지, 개발자가 출처 검토 후 업데이트 필요.

- HTTPS, 요청 시간 제한, 다운로드 크기 제한 적용
- 전체 다운로드·체크섬·manifest ID/버전/최소 앱 버전 검증 후 파일 설치
- SHA-256 검증은 고정한 배포본과의 일치 확인이며 코드 안전성·서명 검증을 의미하지 않음
- Obsidian 앱 버전은 자동 감지하지 않음. 사용자 확인 필요
- 제3자 코드는 활성화 시 Vault 접근·네트워크 호출 가능. 신뢰 여부는 사용자 판단

## 기존 설치 및 실패 처리

- 기존 플러그인의 manifest와 필수 파일 구조가 유효하면 버전과 관계없이 그대로 유지. 다운로드·업그레이드 없음
- 기존 설치본의 무결성·출처는 재검증하지 않음. 인증 설정 `data.json` 내용은 읽지 않음
- 누락 파일·잘못된 manifest·심볼릭 링크·경로 충돌은 오류로 중단
- `community-plugins.json`, 플러그인 `data.json`, 인증·보안 설정은 변경하지 않음
- 네트워크·체크섬 실패 시 Vault·config 초기화 이전에 중단
- 파일 기록 실패 시 리소스 설치기가 이번에 만든 파일 정리. 빈 디렉터리·이미 준비한 Vault/config는 남을 수 있음
- 충돌은 자동 덮어쓰기하지 않음. 동시 설치나 실행 중인 Obsidian의 파일 변경에 대한 격리 보장 없음
- 파일 설치 중에는 Obsidian을 닫는 것을 권장

## 사용자 완료 단계

1. 설치한 OS의 Obsidian에서 Vault 열기 / 재시작
2. 커뮤니티 플러그인 신뢰 안내 확인
3. 설치된 Karpathy LLM Wiki 활성화
4. 플러그인 내부에서 모델 로그인 또는 API 설정 (CLI에 키 전달 금지)
5. 연결 검사 후 `sessions/`의 노트 한 개 ingest
6. `templates/`는 ingest 대상에서 제외

## 검증

- 단위 테스트: 동의 없는 다운로드 방지, checksum/manifest 실패 의미, 기존 설치·설정 보존, 불완전 설치·symlink 거부
- 임시 Vault: 실제 공식 릴리스 다운로드·고정 해시 검증·파일 설치·재실행 보존 확인
- PTY: 플러그인 건너뛰기 후 Obsidian 설정 변경 없이 init 완료 확인
- 실제 Windows 및 macOS Obsidian에서 자동 설치본을 활성화하는 흐름은 별도 검증 필요
- 현재 사용자 Vault는 설치 테스트 대상으로 사용하지 않음
