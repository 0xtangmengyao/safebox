# safebox

[English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md)

AI 사용 가이드: [English](AI_GUIDE.md) | [简体中文](AI_GUIDE.zh-CN.md)

[![CI](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml/badge.svg)](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://img.shields.io/badge/go-1.25%2B-00ADD8.svg)](go.mod)

`safebox`는 로컬 파일과 텍스트를 암호화하기 위한 작고 안전하며 크로스 플랫폼인 Go 명령줄 도구입니다. 애플리케이션 간에 파이프로 암호문을 주고받는 용도에도 적합합니다.

- 성숙한 [age](https://age-encryption.org) 표준 라이브러리(`filippo.io/age`)만 사용합니다. **자체 암호 알고리즘이나 프로토콜을 구현하지 않습니다.**
- 암호문은 기본적으로 age의 **armor 텍스트 형식**(`-----BEGIN AGE ENCRYPTED FILE-----`)으로 출력되어 복사·붙여넣기와 파이프에 편리합니다.
- 기본적으로 stdin/stdout을 읽고 쓰며 POSIX 셸과 PowerShell 모두에서 동작합니다. **평문은 명령줄 인수로 전달할 수 없으므로** 셸 기록과 프로세스 목록에 유출되지 않습니다.
- 순수한 로컬 CLI입니다. **실행 중 네트워크 접근이나 텔레메트리가 전혀 없습니다.** 스스로 인터넷에 연결하지 않습니다.

## 특징

- `filippo.io/age`를 통한 표준 age/X25519 암복호화.
- 스트림 지향이라 큰 파일도 메모리에 올릴 필요가 없습니다.
- 기존 파일 덮어쓰기를 거부하고 `O_EXCL`로 경쟁 상태를 방지합니다.
- 평문을 쓰기 전에 인증 태그를 완전히 검증합니다.
- CLI 도움말은 영어, 중국어 간체, 일본어, 한국어, 스페인어, 프랑스어를 지원합니다.

## 설치 및 빌드

Go 1.25 이상이 필요합니다(현재 `filippo.io/age` 의존성의 최소 버전과 동일).

```sh
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox
go build -o safebox .
```

크로스 컴파일(순수 Go, CGO 불필요):

```sh
# Linux amd64
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
# Windows
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

`$GOBIN` / `%GOPATH%\bin`에 설치:

```sh
go install .
```

### 미리 빌드된 바이너리

`dist/` 디렉터리에는 Linux, macOS, Windows용 amd64·arm64 빌드가 게시되어 있습니다. 운영체제와 CPU 아키텍처에 맞는 파일을 선택하세요. `dist/SHA256SUMS`는 다운로드가 손상되지 않았는지 확인하는 데 사용할 수 있지만, 체크섬만으로 게시자를 증명할 수는 없습니다. 현재 바이너리는 코드 서명되어 있지 않습니다. 높은 민감도의 용도에서는 소스에서 직접 빌드하고 출처를 확인하세요.

버전과 도움말 출력:

```sh
safebox version
safebox help
```

## 명령

### 키 쌍 생성

```sh
safebox keygen --identity ./keys/alice.txt
```

- 개인 키는 `./keys/alice.txt`에 기록되며 POSIX 시스템에서 `0600` 권한을 요청합니다. 공개 키는 `./keys/alice.txt.pub`에 기록되며 `0644` 권한을 요청합니다.
- POSIX 시스템에서 **실제 권한은 프로세스 umask의 제한도 받습니다**. umask는 권한 비트를 제거할 뿐이므로 결과가 요청보다 넓어지지는 않습니다. 예를 들어 umask `022`에서는 공개 키가 `0644`, umask `077`에서는 `0600`이 됩니다. 개인 키는 `0600`을 요청하고 그룹/기타 비트가 없어 일반적으로 umask의 영향을 받지 않습니다.
- Windows에서는 상위 디렉터리에서 상속된 ACL이 파일 접근을 제어합니다. 해당 디렉터리가 의도한 계정에만 접근 가능한지 확인하세요.
- 두 파일 중 하나라도 이미 있으면 `safebox`는 진행을 거부합니다. **절대 덮어쓰지 않습니다.**
- 공개 키(수신자)는 stdout에도 출력되어 배포하기 편리합니다.

### 암호화

기본적으로 stdin에서 평문을 읽고 armor 암호문을 stdout에 씁니다:

```sh
safebox encrypt --recipient age1... < plain.txt
```

수신자를 여러 개 지정할 수 있습니다(일치하는 개인 키 아무거나로 복호화 가능):

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb < plain.txt
```

파일 경로를 사용할 수도 있습니다. 출력 파일은 `0600` 권한으로 생성되며 **덮어쓰지 않습니다**:

```sh
safebox encrypt --recipient age1... --in plain.txt --out secret.age
```

POSIX 시스템에서 요청 출력 권한은 `0600`입니다(마찬가지로 umask로 더 좁아질 수는 있어도 더 넓어지지는 않습니다). Windows에서는 상속된 상위 디렉터리 ACL이 적용됩니다.

> shell의 `> file` 리디렉션은 shell이 처리하며 `safebox`가 시작되기 전에 기존 파일을 잘라낼 수 있습니다. 저장할 때는 `--out PATH`를 사용해 `safebox`의 덮어쓰기 방지 보호를 적용하세요.

### 복호화

기본적으로 stdin에서 암호문을 읽고 평문을 stdout에 씁니다:

```sh
safebox decrypt --identity ./keys/alice.txt < secret.age
```

파일 방식(POSIX 출력 권한 `0600`, Windows 상속 ACL, **덮어쓰지 않음**):

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

아이덴티티 파일은 age 표준 규칙으로 파싱됩니다. age-keygen 형식 파일을 지원하고, 빈 줄과 `#` 주석 줄(`# created: ...`, `# public key: age1...` 등)은 무시되며, 한 파일에 **여러 아이덴티티**를 담을 수 있고 일치하는 것이 있으면 복호화됩니다.

복호화는 먼저 평문을 시스템 임시 디렉터리의 권한이 제한된 임시 파일로 스트리밍하고, 인증 태그를 완전히 검증한 뒤에야 기록합니다. 인증에 실패하면 평문이 전혀 출력되지 않고, 메모리 사용량이 파일 크기에 비례해 늘어나지도 않습니다. 임시 파일은 종료 시 삭제되지만 **삭제는 안전한 소거가 아닙니다**. 임시 디렉터리가 있는 디스크, 스왑, 백업에 데이터가 남을 수 있습니다. 입력은 armor 텍스트와 바이너리 age 형식을 모두 지원합니다.

`--out`이 기존 경로를 가리키면 **복호화나 임시 평문 생성 전에** 거부됩니다. 이 존재 여부 검사는 심볼릭 링크를 따라가지 않으므로 끊어진 심볼릭 링크도 기존 항목으로 간주되어 거부됩니다. 최종 쓰기는 여전히 `O_EXCL`을 경쟁 방지의 최종 방어선으로 사용합니다. 기존 `--out` 대상은 어떤 경우에도 덮어쓰거나 삭제하지 않습니다. 다만 이번 실행에서 새 파일을 만든 뒤 쓰기 실패가 발생하면 이번에 생성된 부분 출력은 정리됩니다.

## 예제

### 파이프로 텍스트 전달(Linux / macOS / Git Bash)

```sh
printf 'hello safebox\n' | safebox encrypt --recipient AGE1... | safebox decrypt --identity alice.txt
```

### 파일

```sh
safebox keygen --identity alice.txt
safebox encrypt --recipient "$(cat alice.txt.pub)" --in report.pdf --out report.pdf.age
safebox decrypt --identity alice.txt --in report.pdf.age --out report.pdf
```

### 애플리케이션 간 전달

```sh
# 애플리케이션 A가 파이프로 암호문을 출력
producer-app | safebox encrypt --recipient age1... | consumer-app
```

### Windows PowerShell

```powershell
# 키 쌍 생성
safebox keygen --identity .\keys\alice.txt

# 파일 모드는 PowerShell의 텍스트 인코딩/줄바꿈 변환을 피할 수 있음
safebox encrypt --recipient (Get-Content -Raw .\keys\alice.txt.pub).Trim() `
    --in .\plain.txt --out .\secret.age

# 복호화
safebox decrypt --identity .\keys\alice.txt `
    --in .\secret.age --out .\plain.out.txt
```

> 참고: 파일 모드는 PowerShell의 텍스트 개체 파이프라인을 거치지 않으므로 임의 파일에 적합합니다. `echo`/`printf` 명령 내용으로 민감한 텍스트를 직접 입력하면 셸 기록에 남을 수 있습니다. 민감한 텍스트는 파일, 클립보드 도구 또는 상위 애플리케이션에서 stdin으로 전달하세요. 평문이나 개인 키를 `safebox` 명령줄에 넣지 마세요.

## 언어 및 현지화

CLI 도움말과 진단은 영어(`en`), 중국어 간체(`zh-CN`), 일본어(`ja`), 한국어(`ko`), 스페인어(`es`), 프랑스어(`fr`)를 지원합니다. 기본값은 영어입니다.

전역 `--lang` 플래그 또는 `SAFEBOX_LANG` 환경 변수로 명시적으로 전환할 수 있습니다:

```sh
safebox --lang ja help      # 일본어 도움말
SAFEBOX_LANG=fr safebox decrypt   # 프랑스어 진단
```

`--lang`은 `SAFEBOX_LANG`보다 우선하며, 알 수 없는 값은 영어로 대체됩니다. 자신의 AI 어시스턴트가 `safebox`를 안전하게 사용하도록 하는 방법은 [AI 사용 가이드](AI_GUIDE.md)([简体中文](AI_GUIDE.zh-CN.md))를 참고하세요.

## 보안 경계

- **암호 기술**: `filippo.io/age`를 통한 표준 age / X25519. `safebox`는 자체 암호를 발명하지 않습니다.
- **개인 키 보호**: POSIX 시스템에서 개인 키 파일은 `0600`으로 생성되고, Windows에서는 상위 디렉터리 ACL의 적용을 받습니다. `safebox`는 개인 키를 절대 출력하지 않습니다. 키 파일을 스스로 잘 보호하고 공용 또는 비암호화 디스크에 두지 마세요.
- **임시 평문**: 복호화는 전체 스트림을 검증하기 위해 권한이 제한된 임시 파일을 사용합니다. 종료 후 삭제되지만 SSD, 스왑, 백업, 스냅샷에서 안전하게 소거된다는 보장은 없습니다. 높은 기밀성이 필요하면 임시 디렉터리와 스왑을 신뢰할 수 있는 암호화 저장소에 두세요.
- **복구 불가**: 개인 키를 잃으면 암호문은 **복구할 수 없습니다**. 백도어도 복구 메커니즘도 없습니다.
- **메타데이터**: 암호화는 파일 내용만 보호합니다. age는 파일 크기, 타임스탬프, 파일 이름을 숨기지 않으며 트래픽 패턴도 숨기지 않습니다.
- **터미널과 클립보드**: 터미널을 통해 키나 평문을 복사·붙여넣으면 터미널 기록, 스크롤백 또는 클립보드에 남을 수 있습니다.
- **패스프레이즈 엔트로피**: 이 도구는 공개 키 암호를 사용하며 패스프레이즈에 의존하지 않습니다. 패스프레이즈 암호화에는 낮은 엔트로피의 암호를 사용하지 마세요.
- **로컬 도구**: `safebox`는 로컬 CLI이며 실행 중 네트워크 기능이나 텔레메트리가 전혀 없습니다.

## 개발 및 테스트

```sh
gofmt -l .
go vet ./...
go test ./...
```

테스트는 키 생성과 파일 권한(플랫폼별 조정), stdin/stdout 및 파일 왕복, 다중 수신자, 암호문 변조나 잘못된 키 사용 시 평문 누출 없는 실패, 기존 파일 덮어쓰기 거부를 다룹니다. 또한 배포되는 모든 언어의 번역 완전성도 검증합니다.

지속적 통합은 모든 push와 pull request에서 빌드와 테스트를 수행하고 `gofmt`, `go vet`, `go test`를 실행하며 Linux, macOS, Windows용으로 크로스 컴파일합니다(산출물은 저장소에 남기지 않습니다).

## 프로젝트 메타데이터(제안)

아래 GitHub topics는 제안이며 저장소 설정에서 직접 지정할 수 있습니다. 이 저장소는 온라인 메타데이터를 자동으로 변경하지 않습니다:

`age`, `encryption`, `cryptography`, `x25519`, `cli`, `golang`, `privacy`, `command-line-tool`, `file-encryption`, `cross-platform`

## 기여

빌드, 테스트, 변경 제출 방법은 [CONTRIBUTING.md](CONTRIBUTING.md)를 참고하세요. 취약점은 [SECURITY.md](SECURITY.md)에 설명된 방식으로 비공개로 신고해 주세요. 이 프로젝트는 [Contributor Covenant](CODE_OF_CONDUCT.md)를 따릅니다.

## 라이선스

MIT. [LICENSE](LICENSE)를 참고하세요.
