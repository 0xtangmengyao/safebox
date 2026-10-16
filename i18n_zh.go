package main

// zhMessages is the Simplified Chinese catalog.
var zhMessages = catalog{
	// Global usage.
	"usage.title":       "safebox %s - 基于 age 的文件与文本加密",
	"usage.header":      "用法：",
	"usage.synopsis":    "  safebox [--lang CODE] <命令> [参数]",
	"usage.commands":    "命令：",
	"usage.keygen":      "  keygen   生成 X25519 身份（私钥）及其公钥收件人",
	"usage.encrypt":     "  encrypt  将 stdin（或 --in）加密为 armor 文本输出到 stdout（或 --out）",
	"usage.decrypt":     "  decrypt  将 armor/二进制 age 输入解密到 stdout（或 --out）",
	"usage.version":     "  version  打印版本",
	"usage.help":        "  help     打印本帮助",
	"usage.commandHelp": `运行 "safebox <命令> -h" 查看该命令的参数。`,
	"usage.plaintext": "明文永远不会作为命令行参数传入：请通过 stdin 管道输入，\n" +
		"  以免泄露到 shell 历史记录和进程列表。",
	"usage.lang": "语言：使用 --lang en|zh-CN|ja|ko|es|fr 或环境变量 SAFEBOX_LANG。",

	// Diagnostics.
	"diag.missingCommand":            "缺少命令",
	"diag.unknownCommand":            "未知命令 %q",
	"diag.unexpectedArgument":        "%s：多余的参数 %q",
	"diag.keygen.requiresIdentity":   "keygen 需要 --identity PATH",
	"diag.encrypt.requiresRecipient": "encrypt 至少需要一个 --recipient",
	"diag.decrypt.requiresIdentity":  "decrypt 需要 --identity PATH",
	"diag.interrupted":               "已中断",

	// Per-command usage.
	"keygen.usageLine": "用法：safebox keygen --identity PATH",
	"keygen.detail": "写入 PATH（私钥，0600）和 PATH.pub（公钥，0644）。\n" +
		"两个文件都绝不覆盖。",

	"encrypt.usageLine": "用法：safebox encrypt --recipient AGE1... [--recipient ...] [--in PATH] [--out PATH]",
	"encrypt.detail": "未指定 --in 时从 stdin 读取明文，输出 armor 文本。\n" +
		"明文不得作为参数传入。",

	"decrypt.usageLine": "用法：safebox decrypt --identity PATH [--in PATH] [--out PATH]",
	"decrypt.detail": "接受 armor 或二进制 age 输入。明文会在写出任何内容之前\n" +
		"完成完整校验，因此失败时绝不会泄露部分输出。",

	// Flag descriptions.
	"flag.keygen.identity":   "要创建的 X25519 私钥文件路径（0600）",
	"flag.encrypt.recipient": "收件人公钥（age1...）；可重复指定",
	"flag.encrypt.in":        "输入文件（默认 stdin）",
	"flag.encrypt.out":       "输出文件（默认 stdout）；绝不覆盖",
	"flag.decrypt.identity":  "X25519 私钥文件路径",
	"flag.decrypt.in":        "输入文件（默认 stdin）",
	"flag.decrypt.out":       "输出文件（默认 stdout）；绝不覆盖",
}
