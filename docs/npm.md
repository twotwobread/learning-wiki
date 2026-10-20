# npm 패키지 (Linux x64 / macOS arm64)

## 상태

- 메인 패키지 `@suyoung9137/learning-wiki`와 대상별 바이너리 패키지 `@suyoung9137/learning-wiki-linux-x64`, `@suyoung9137/learning-wiki-darwin-arm64` 사용
- npm 계정: `suyoung9137`
- 자체 코드: MIT 라이선스. 번들 의존성은 해당 라이선스와 고지 적용
- `0.2.0`은 두 플랫폼 패키지와 메인 패키지를 같은 버전으로 발행 완료. 개발용 루트 package.json만 `private: true` 유지
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

## 태그 기반 배포 (`0.2.0` 이후)

`0.2.0` 수동 발행은 완료했습니다. 이후 버전은 세 패키지의 버전을 함께 올려 `main`에 반영하고, 해당 커밋에 `vX.Y.Z` 태그를 push합니다. `.github/workflows/release.yml`은 정확한 안정 버전 태그·패키지 메타데이터·깨끗한 checkout·`origin/main` 조상을 확인합니다. Linux에서 한 번 빌드한 세 tarball과 `artifacts.json`을 하나의 artifact로 업로드하고, Linux x64와 macOS arm64에서 같은 artifact를 검증합니다. 두 smoke 작업이 성공하면 `npm-release` Environment 승인 후 같은 파일을 OIDC로 발행합니다.

### 운영 설정 (첫 자동 배포 직전)

워크플로 파일만으로 승인 대기가 보장되지는 않습니다. 현재 `npm-release` Environment는 생성되지 않았습니다. 아래 GitHub/npm 설정은 운영자가 별도로 완료해야 하며, 이 문서나 워크플로 추가로 자동 설정되지 않습니다.

1. GitHub 저장소 `twotwobread/learning-wiki`의 Settings → Environments에서 `npm-release`를 생성합니다. Required reviewers에 배포 승인 담당자 `twotwobread`를 지정합니다. 단독 관리자가 자신의 태그 배포를 승인할 수 있도록 **Prevent self-review는 끕니다**. 별도 승인 담당자를 지정한 경우에는 Prevent self-review를 켤 수 있습니다. 관리자의 보호 규칙 우회는 허용하지 않습니다.
2. Deployment branches and tags를 **Selected branches and tags**로 설정하고 **Tag** 유형의 `v*` 규칙만 추가합니다. Branch `main` 규칙만으로는 태그 배포를 허용할 수 없습니다. 태그 커밋이 `main`에 포함되는지는 워크플로의 조상 검사가 보장합니다. GitHub 요금제·저장소 가시성에서 Required reviewers 기능을 사용할 수 있는지도 확인합니다.
3. npm에서 아래 세 패키지 각각의 Settings → Trusted Publisher에 GitHub Actions를 등록합니다. Organization or user는 `twotwobread`, Repository는 `learning-wiki`, Workflow filename은 **`release.yml`**, Environment는 **`npm-release`**로 정확히 입력합니다. 직접 `npm publish` 가능한 권한으로 설정합니다.

   - `@suyoung9137/learning-wiki-linux-x64`
   - `@suyoung9137/learning-wiki-darwin-arm64`
   - `@suyoung9137/learning-wiki`

새 npm Trusted Publisher는 생성 후 **2일 이내 첫 발행**을 성공해야 합니다. 첫 자동 배포 직전에 세 설정을 생성하고, 기한이 지나면 npm 설정을 확인해 다시 등록합니다. npm 토큰을 GitHub Secrets나 저장소에 추가하지 않습니다. publish 작업은 `contents: read`, `id-token: write`만 사용하며 Node `22.23.3`과 npm CLI `11.21.0`을 설치합니다. 공개 저장소·GitHub 호스팅 runner 등 npm 조건을 충족하면 provenance가 첨부됩니다.

### 승인과 부분 실패 재시도

승인자는 태그·커밋·테스트 결과·artifact를 확인한 뒤 Environment 배포를 승인합니다. 플랫폼 패키지 두 개를 먼저, 메인 패키지를 마지막으로 발행합니다. 각 발행 후 레지스트리의 `dist.integrity`를 확인하고, 마지막에 세 패키지를 다시 확인합니다.

일부만 발행된 경우 GitHub Actions에서 해당 실행의 **실패한 작업만 재실행**하여 같은 artifact를 사용합니다. 승인 단계가 다시 나타나면 다시 승인합니다. 이미 존재하는 이름/버전은 검증된 tarball과 registry integrity가 같을 때만 건너뛰며, 나머지 패키지만 순서대로 발행합니다. integrity가 다르면 중단하고 원인을 조사한 뒤 새 버전을 사용합니다. 같은 버전은 덮어쓸 수 없습니다. artifact가 만료되거나 삭제되었으면 원본 artifact를 확보하기 전까지 재시도하지 않습니다.

전역 concurrency 그룹은 배포 실행을 직렬화하고 진행 중인 발행을 취소하지 않습니다. GitHub concurrency는 대기 실행을 하나만 유지하므로, 승인 대기 중 여러 태그를 연속 push하면 이전 대기 실행이 대체될 수 있습니다. 한 릴리스가 끝난 뒤 다음 태그를 push합니다.

로컬 tarball 검증, 네이티브 runner 검증, 실제 npm 설치, Obsidian 앱·동기화 연동은 각각 구분합니다. 기존 설치 프로그램의 사용자 권한, 경로·파일 보존, 플러그인 동의 정책은 npm 래퍼에서도 동일하게 적용합니다.
