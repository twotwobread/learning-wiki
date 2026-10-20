# npm 패키지 (Linux x64 / macOS arm64)

## 상태

- 메인 패키지 `@suyoung9137/learning-wiki`와 대상별 바이너리 패키지 `@suyoung9137/learning-wiki-linux-x64`, `@suyoung9137/learning-wiki-darwin-arm64` 사용
- npm 계정: `suyoung9137`
- 자체 코드: MIT 라이선스. 번들 의존성은 해당 라이선스와 고지 적용
- `0.2.0`은 두 플랫폼 패키지와 메인 패키지를 같은 버전으로 발행. 개발용 루트 package.json만 `private: true` 유지
- 레지스트리 제공 버전 확인: `npm view @suyoung9137/learning-wiki version`

## 레지스트리 설치

기존 `0.1.0`은 WSL Linux x64에서 실행:

```bash
npx @suyoung9137/learning-wiki@0.1.0 --version
npx @suyoung9137/learning-wiki@0.1.0 init
```

Node.js 22 이상 필요. Go·mise·저장소 clone은 최종 사용자에게 불필요.

`0.2.0`은 WSL Linux x64와 macOS arm64에서 같은 메인 명령을 사용합니다:

```bash
npx @suyoung9137/learning-wiki@0.2.0 init
```

이미 발행된 `0.2.0`에서는 수집 스킬 설치에 에이전트 선택이 필요합니다. 대화형 마법사에서 Space로 pi 또는 Codex를 선택하거나, 비대화형으로 명시합니다:

```bash
npx @suyoung9137/learning-wiki@0.2.0 init --vault "$HOME/Documents/LearningWiki" --create --agents codex
```

이 브랜치의 변경은 다음 릴리스에 반영됩니다. `init --vault`는 공통 스킬을 기본 설치하고 `--no-skill`로 생략합니다. 대화형 마법사에서는 pi·Codex를 기본 선택하며 모두 해제하면 스킬 설치를 생략합니다. Codex는 `~/.agents/skills/learning-capture`에서 스킬을 탐색하며 발견되지 않으면 재시작 후 확인합니다. 발행된 `0.2.0`의 동작은 바뀌지 않습니다.

다음 릴리스 발행 전 검증에는 아래 로컬 패키징 절차를 사용합니다.

## 구조

```text
메인 npm 패키지: 작은 JS 래퍼
    ↓ optionalDependencies (동일 버전 고정)
Linux x64 또는 macOS arm64 패키지: Go 실행 파일 + 의존성 라이선스 고지
```

- 런처는 플랫폼 검사, 바이너리 경로 탐색, 인자·표준 입출력·종료 코드 전달만 담당
- 별도 postinstall 다운로드 없음. 플랫폼 패키지를 npm이 선택
- 지원 대상은 WSL Linux x64와 네이티브 macOS arm64. Windows PowerShell에서 실행하지 않기
- Node.js 22 이상 필요. 사용자 환경에는 Go나 mise 불필요
- optional dependencies를 제외하면 실행 불가; 해당 오류에서 `--include=optional` 안내
- `bin/targets.json`에서 npm 패키지와 `GOOS/GOARCH` 조합을 선택. 빌드 시 `CGO_ENABLED=0` 사용
- 플러그인은 패키지에 번들하지 않음. 사용자가 init에서 동의한 경우 별도 다운로드

## 개발 및 로컬 패키징

저장소 루트에서:

```bash
mise install
mise exec -- go mod download
mise exec -- npm test
mise exec -- npm run pack:local
mise exec -- npm run test:package
```

- `dist/npm/`: 생성된 패키지 디렉터리
- `dist/*.tgz`: 설치 가능한 npm tarball
- `dist/artifacts.json`: 패키지명·버전·파일명·npm integrity
- 바이너리 버전과 npm 패키지 버전 일치 확인
- 패키징 시 기존 `dist/npm/` 스테이징만 재생성
- 다른 버전의 이전 tarball은 남을 수 있음. 테스트는 artifacts.json에 기록된 현재 결과만 사용

수동 로컬 설치 예시 (테스트용 빈 디렉터리에서):

```bash
target="$(node -p 'process.platform + "-" + process.arch')"
npm install --offline --ignore-scripts --no-audit --no-fund \
  "/absolute/path/to/dist/suyoung9137-learning-wiki-${target}-0.2.0.tgz" \
  /absolute/path/to/dist/suyoung9137-learning-wiki-0.2.0.tgz
npx --offline --no-install learning-wiki --version
npx --offline --no-install learning-wiki init
```

`target`은 WSL Linux x64에서 `linux-x64`, Mac ARM64에서 `darwin-arm64`입니다. 오프라인 로컬 검증에는 메인과 현재 OS용 tarball 두 개만 지정합니다. 레지스트리 배포 후 일반 설치에서는 npm이 동일 버전의 대상 패키지를 선택합니다.

## 검증 범위

- Node 단위 테스트: 미지원 플랫폼·플랫폼 패키지 누락 오류, 인자 보존, 출력 스트림·종료 코드 전달, SIGTERM 전달
- 임시 HOME와 빈 npm 캐시: 현재 OS용 tarball의 offline 설치 → npx 버전 출력 → 잘못된 입력 → 한글 Vault 초기화·재실행 검증
- PTY 수동 smoke test: 설치된 패키지의 npx → 마법사 표시 → Ctrl+C 종료 코드 130, config 미생성
- `.github/workflows/ci.yml`: Linux x64와 macOS arm64에서 Go/Node 테스트·race 검사·로컬 패키징 검증 및 tarball artifact 업로드. publish 단계 없음
- 로컬 tarball 검증은 실제 npm 레지스트리 설치·Obsidian ingest 검증과 구분

## 배포 전 확인

1. npm 계정·scope 소유권 및 패키지명 확정
2. 자체 코드 라이선스 결정, 바이너리 의존성 고지 검토
3. 공개 배포 승인 확인
4. CI에서 Go/Node 테스트·빌드·패키징 검증
5. 실제 WSL Linux x64와 macOS arm64에서 각각 검증
6. 명시적 배포 승인 후 플랫폼 패키지 먼저, 메인 패키지 다음 순서로 동일 버전 배포

배포 명령 (npm 로그인 및 필요 시 웹/OTP 인증):

```bash
npm publish ./dist/suyoung9137-learning-wiki-linux-x64-0.2.0.tgz --access public
npm publish ./dist/suyoung9137-learning-wiki-darwin-arm64-0.2.0.tgz --access public
npm publish ./dist/suyoung9137-learning-wiki-0.2.0.tgz --access public
```

로컬 tarball 경로는 `./dist/...` 또는 절대 경로로 지정 (일부 npm 버전에서 `dist/...`를 GitHub 경로로 해석 가능).

로그인 성공이 배포 권한 충족을 의미하지는 않음. npm이 2FA를 요구하면 계정 보안 설정을 완료하고 사용자 터미널에서 인증 진행. 인증정보·OTP를 저장소나 채팅에 기록하지 않기. 같은 버전은 재배포 불가하므로 검증한 tarball만 사용.

기존 설치 프로그램의 사용자 권한, 경로·파일 보존, 플러그인 동의 정책은 npm 래퍼에서도 동일하게 적용.
