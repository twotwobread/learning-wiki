# npm 배포 준비 (Linux x64 / WSL)

## 상태

- 패키지명은 잠정적으로 `@twotwobread/learning-wiki`, `@twotwobread/learning-wiki-linux-x64` 사용
- 실제 npm 계정·scope 권한 미확인, 레지스트리 미배포
- 우발적 배포 방지를 위해 두 패키지 모두 `private: true`
- 자체 코드의 배포 라이선스 결정 전까지 `UNLICENSED`. 공개 배포 전 권한·라이선스·private 설정 검토 필요
- 지금은 로컬 tarball로 설치·실행 가능. `npx @twotwobread/learning-wiki`의 레지스트리 설치는 아직 사용 불가

## 구조

```text
메인 npm 패키지: 작은 JS 래퍼
    ↓ optionalDependencies (동일 버전 고정)
Linux x64 패키지: Go 실행 파일 + 의존성 라이선스 고지
```

- 런처는 플랫폼 검사, 바이너리 경로 탐색, 인자·표준 입출력·종료 코드 전달만 담당
- 별도 postinstall 다운로드 없음. 플랫폼 패키지를 npm이 선택
- 첫 지원 대상은 WSL Linux x64. Windows PowerShell에서 실행하지 않기
- Node.js 22 이상 필요. 사용자 환경에는 Go나 mise 불필요
- optional dependencies를 제외하면 실행 불가; 해당 오류에서 `--include=optional` 안내
- Go 바이너리는 `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`로 빌드
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
npm install --offline --ignore-scripts --no-audit --no-fund \
  /absolute/path/to/dist/twotwobread-learning-wiki-linux-x64-0.1.0.tgz \
  /absolute/path/to/dist/twotwobread-learning-wiki-0.1.0.tgz
npx --offline --no-install learning-wiki --version
npx --offline --no-install learning-wiki init
```

설치 시에는 두 로컬 tarball을 함께 지정. 아직 레지스트리에 없으므로 메인 tarball만으로 플랫폼 패키지 다운로드 불가.

## 검증 범위

- Node 단위 테스트: 미지원 플랫폼·플랫폼 패키지 누락 오류, 인자 보존, 출력 스트림·종료 코드 전달, SIGTERM 전달
- 임시 HOME와 빈 npm 캐시: 두 tarball의 offline 설치 → npx 버전 출력 → 잘못된 입력 → 한글 Vault 초기화·재실행 검증
- PTY 수동 smoke test: 설치된 패키지의 npx → 마법사 표시 → Ctrl+C 종료 코드 130, config 미생성
- `.github/workflows/ci.yml`: mise 기반 Go/Node 테스트·race 검사·로컬 패키징 검증 및 tarball artifact 업로드. publish 단계 없음
- 위 로컬 검증은 현재 WSL 환경에서 수행. 다른 기기·실제 npm 레지스트리 설치와 동일한 검증으로 간주하지 않음

## 배포 전 확인

1. npm 계정·scope 소유권 및 패키지명 확정
2. 자체 코드 라이선스 결정, 바이너리 의존성 고지 검토
3. private 보호 해제 여부 승인
4. CI에서 Go/Node 테스트·빌드·패키징 검증
5. 실제 다른 WSL 환경에서 검증
6. 명시적 배포 승인 후 플랫폼 패키지 먼저, 메인 패키지 다음 순서로 동일 버전 배포

기존 설치 프로그램의 사용자 권한, 경로·파일 보존, 플러그인 동의 정책은 npm 래퍼에서도 동일하게 적용.
