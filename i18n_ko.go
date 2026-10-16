package main

// koMessages is the Korean catalog.
var koMessages = catalog{
	// Global usage.
	"usage.title":       "safebox %s - age 기반 파일 및 텍스트 암호화",
	"usage.header":      "사용법:",
	"usage.synopsis":    "  safebox [--lang CODE] <명령> [플래그]",
	"usage.commands":    "명령:",
	"usage.keygen":      "  keygen   X25519 아이덴티티와 해당 공개 수신자를 생성합니다",
	"usage.encrypt":     "  encrypt  stdin(또는 --in)을 armor 텍스트로 stdout(또는 --out)에 암호화합니다",
	"usage.decrypt":     "  decrypt  armor/바이너리 age 입력을 stdout(또는 --out)으로 복호화합니다",
	"usage.version":     "  version  버전을 출력합니다",
	"usage.help":        "  help     이 도움말을 출력합니다",
	"usage.commandHelp": `명령별 플래그는 "safebox <명령> -h"를 실행하세요.`,
	"usage.plaintext": "평문은 명령줄 인수로 절대 전달할 수 없습니다. 셸 기록과 프로세스 목록\n" +
		"  유출을 피하려면 stdin으로 파이프하세요.",
	"usage.lang": "언어: --lang en|zh-CN|ja|ko|es|fr 또는 환경 변수 SAFEBOX_LANG을 사용하세요.",

	// Diagnostics.
	"diag.missingCommand":            "명령이 없습니다",
	"diag.unknownCommand":            "알 수 없는 명령 %q",
	"diag.unexpectedArgument":        "%s: 예상치 못한 인수 %q",
	"diag.keygen.requiresIdentity":   "keygen에는 --identity PATH가 필요합니다",
	"diag.encrypt.requiresRecipient": "encrypt에는 최소 하나의 --recipient가 필요합니다",
	"diag.decrypt.requiresIdentity":  "decrypt에는 --identity PATH가 필요합니다",
	"diag.interrupted":               "중단됨",

	// Per-command usage.
	"keygen.usageLine": "사용법: safebox keygen --identity PATH",
	"keygen.detail": "PATH(개인 키, 0600)와 PATH.pub(공개 키, 0644)를 씁니다.\n" +
		"두 파일 모두 절대 덮어쓰지 않습니다.",

	"encrypt.usageLine": "사용법: safebox encrypt --recipient AGE1... [--recipient ...] [--in PATH] [--out PATH]",
	"encrypt.detail": "--in이 없으면 stdin에서 평문을 읽고 armor 텍스트를 출력합니다.\n" +
		"평문을 인수로 전달해서는 안 됩니다.",

	"decrypt.usageLine": "사용법: safebox decrypt --identity PATH [--in PATH] [--out PATH]",
	"decrypt.detail": "armor 또는 바이너리 age 입력을 받습니다. 평문은 출력 전에\n" +
		"완전히 검증되므로 실패 시 부분 출력이 새지 않습니다.",

	// Flag descriptions.
	"flag.keygen.identity":   "생성할 X25519 개인 키 파일의 경로(0600)",
	"flag.encrypt.recipient": "수신자 공개 키(age1...); 여러 번 지정할 수 있습니다",
	"flag.encrypt.in":        "입력 파일(기본값 stdin)",
	"flag.encrypt.out":       "출력 파일(기본값 stdout); 덮어쓰지 않습니다",
	"flag.decrypt.identity":  "X25519 개인 키 파일의 경로",
	"flag.decrypt.in":        "입력 파일(기본값 stdin)",
	"flag.decrypt.out":       "출력 파일(기본값 stdout); 덮어쓰지 않습니다",
}
