package main

// jaMessages is the Japanese catalog.
var jaMessages = catalog{
	// Global usage.
	"usage.title":       "safebox %s - age ベースのファイル・テキスト暗号化",
	"usage.header":      "使い方:",
	"usage.synopsis":    "  safebox [--lang CODE] <コマンド> [フラグ]",
	"usage.commands":    "コマンド:",
	"usage.keygen":      "  keygen   X25519 秘密鍵とその公開鍵（受信者）を生成します",
	"usage.encrypt":     "  encrypt  stdin（または --in）を armor テキストとして stdout（または --out）へ暗号化します",
	"usage.decrypt":     "  decrypt  armor/バイナリ age 入力から stdout（または --out）へ復号します",
	"usage.version":     "  version  バージョンを表示します",
	"usage.help":        "  help     このヘルプを表示します",
	"usage.commandHelp": `コマンド固有のフラグは "safebox <コマンド> -h" を実行してください。`,
	"usage.plaintext": "平文をコマンドライン引数として渡すことはできません。shell の履歴や\n" +
		"  プロセス一覧への漏えいを避けるため、stdin からパイプで入力してください。",
	"usage.lang": "言語: --lang en|zh-CN|ja|ko|es|fr または環境変数 SAFEBOX_LANG を使用します。",

	// Diagnostics.
	"diag.missingCommand":            "コマンドがありません",
	"diag.unknownCommand":            "不明なコマンド %q",
	"diag.unexpectedArgument":        "%s: 予期しない引数 %q",
	"diag.keygen.requiresIdentity":   "keygen には --identity PATH が必要です",
	"diag.encrypt.requiresRecipient": "encrypt には少なくとも 1 つの --recipient が必要です",
	"diag.decrypt.requiresIdentity":  "decrypt には --identity PATH が必要です",
	"diag.interrupted":               "中断されました",

	// Per-command usage.
	"keygen.usageLine": "使い方: safebox keygen --identity PATH",
	"keygen.detail": "PATH（秘密鍵、0600）と PATH.pub（公開鍵、0644）を書き込みます。\n" +
		"どちらのファイルも上書きしません。",

	"encrypt.usageLine": "使い方: safebox encrypt --recipient AGE1... [--recipient ...] [--in PATH] [--out PATH]",
	"encrypt.detail": "--in を指定しない限り stdin から平文を読み取り、armor テキストを出力します。\n" +
		"平文を引数として渡さないでください。",

	"decrypt.usageLine": "使い方: safebox decrypt --identity PATH [--in PATH] [--out PATH]",
	"decrypt.detail": "armor またはバイナリの age 入力を受け付けます。平文は書き出す前に\n" +
		"完全に検証されるため、失敗時に部分的な出力が漏れることはありません。",

	// Flag descriptions.
	"flag.keygen.identity":   "作成する X25519 秘密鍵ファイルのパス（0600）",
	"flag.encrypt.recipient": "受信者の公開鍵（age1...）。複数回指定できます",
	"flag.encrypt.in":        "入力ファイル（既定は stdin）",
	"flag.encrypt.out":       "出力ファイル（既定は stdout）。上書きしません",
	"flag.decrypt.identity":  "X25519 秘密鍵ファイルのパス",
	"flag.decrypt.in":        "入力ファイル（既定は stdin）",
	"flag.decrypt.out":       "出力ファイル（既定は stdout）。上書きしません",
}
