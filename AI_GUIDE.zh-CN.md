# safebox AI 使用指南

[English](AI_GUIDE.md) | [简体中文](AI_GUIDE.zh-CN.md)

本文档用于**供用户自己的 AI 助手阅读或直接粘贴给 AI**。它让 AI 在了解
`safebox` 的足够信息后，能够解释工具用法并起草命令，同时**绝不接触你的任何敏感信息**。

## 请先阅读：秘密信息与 AI

> **不要把秘密信息交给 AI，也不要让 AI 向你索取秘密信息。**
>
> 绝不向任何 AI 助手、聊天机器人或在线服务提供、粘贴、上传或口述以下内容：
>
> - 私钥或身份文件（例如 `AGE-SECRET-KEY-1...` 开头的 age 私钥，或 `./keys/alice.txt` 的内容）。
> - 你想要保护的明文。
> - 你视为敏感的密文（`*.age` 文件，或 `-----BEGIN AGE ENCRYPTED FILE-----` 开头的 armor 文本块）。
> - 口令、API key、token 或任何其他凭据。
>
> AI **不得**向你索取私钥、明文、密文、口令或 token，也不得要求你展示或粘贴任何本地密钥、输入或输出文件的内容。

AI 使用本指南时必须遵守：

1. 仅依据本文档以及你提供的**非敏感描述**作答（例如“我在 `./keys/alice.txt` 有一把私钥”——只说**路径**，绝不说**内容**）。
2. 只给出说明与**命令草案**，不得声称自己执行过任何操作。
3. **不得执行命令。** 即使 AI 拥有 shell 或其他工具，也不得为此任务使用；只能输出说明和命令草案。
4. **不得打开、读取、解密、加密或上传任何本地密钥、输入或输出文件。**
5. 占位符按字面处理：`PATH`、`AGE1...`、`plain.txt`、`secret.age` 等都需要你自己替换。
6. 所有命令都必须由**你本人审阅后亲自执行**。

> 本指南**不是**对各类 AI 的技术控制手段，无法强制 AI、插件或托管服务的行为，也无法让它遗忘数据。你有责任遵守你所选 AI 服务的隐私政策与数据处理规则，并自行判断哪些内容可以与其分享。

## safebox 是什么

`safebox` 是一个小巧、安全、可跨平台的 Go 命令行工具，用于加密本地文件与文本，也适合在应用之间通过管道传输密文。

- 只使用成熟的 [age](https://age-encryption.org) 标准库（`filippo.io/age`），**不实现任何自定义加密算法**。
- 使用 age 的 **X25519** 公钥方案。
- 密文默认输出为 age 的 **armor 文本格式**（`-----BEGIN AGE ENCRYPTED FILE-----`）；解密时也能读取二进制 age 输入。
- 默认读写 stdin/stdout，同时支持 POSIX shell 与 PowerShell。
- **明文不能作为命令行参数传入**，避免泄露到 shell 历史和进程列表。
- 这是一个纯本地 CLI：**没有任何运行时网络或遥测行为**，不会自行连接互联网，也不调用任何 AI 或 HTTP 接口。

`safebox` 没有 `ask` 命令，没有内置 AI，也没有任何网络功能。AI 协助完全来自**你自己的**助手，与这个工具本身无关。

## 环境要求

- Go **1.25 或更高版本**（与当前 `filippo.io/age` 依赖的最低版本一致）。
- 无需 CGO。

## 克隆、构建与安装

```sh
# 克隆
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox

# 构建本地二进制
go build -o safebox .

# 安装到 $GOBIN / %GOPATH%\bin
go install .
```

交叉编译（纯 Go，无需 CGO）：

```sh
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

检查构建：

```sh
safebox version
safebox help
```

## 支持的语言

CLI 帮助与诊断支持：

| 代码    | 语言           |
| ------- | -------------- |
| `en`    | 英语（默认）   |
| `zh-CN` | 简体中文       |
| `ja`    | 日语           |
| `ko`    | 韩语           |
| `es`    | 西班牙语       |
| `fr`    | 法语           |

可通过全局 `--lang` 参数（放在命令之前）或 `SAFEBOX_LANG` 环境变量选择语言。`--lang` 优先，无法识别的值回退为英语。

```sh
safebox --lang ja help
SAFEBOX_LANG=fr safebox decrypt --identity PATH
```

## 命令

所有命令都支持 `-h` / `--help`。

### 生成密钥对（`keygen`）

```sh
safebox keygen --identity PATH
```

- 私钥写入 `PATH`（POSIX 下请求权限 `0600`），公钥写入 `PATH.pub`（请求权限 `0644`）。
- POSIX 下实际权限还受进程 umask 限制，因此绝不会比请求值更宽。
- Windows 下由父目录继承的 ACL 控制访问。
- **绝不覆盖。** 任一文件已存在时都会拒绝继续。
- 公钥（收件人，`age1...`）会打印到 stdout。

使用明确占位符路径的示例：

```sh
safebox keygen --identity ./keys/alice.txt
# 私钥:   ./keys/alice.txt      （预期 0600）
# 公钥:   ./keys/alice.txt.pub  （预期 0644）
# 收件人: age1...               （同时打印到 stdout）
```

### 加密（`encrypt`）

```sh
safebox encrypt --recipient AGE1... [--recipient AGE1...] [--in PATH] [--out PATH]
```

- 未指定 `--in PATH` 时从 stdin 读入明文。
- 未指定 `--out PATH` 时向 stdout 输出 armor 密文。
- 可重复使用 `--recipient` 指定**多个收件人**；任意一个匹配的私钥都能解密。
- 输出文件以 `0600` 创建（POSIX，受 umask 限制），且**绝不覆盖**。
- 明文**不能**作为参数传入。

stdin/stdout 示例：

```sh
printf 'hello safebox\n' | safebox encrypt --recipient AGE1...
```

文件示例：

```sh
safebox encrypt --recipient AGE1... --in plain.txt --out secret.age
```

> shell 的 `> 文件` 重定向由 shell 处理，可能在 `safebox` 启动前截断已有文件。需要保存输出时优先使用 `--out PATH`，以便 `safebox` 拒绝覆盖已存在目标。

多个收件人：

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb --in plain.txt --out secret.age
```

### 解密（`decrypt`）

```sh
safebox decrypt --identity PATH [--in PATH] [--out PATH]
```

- 未指定 `--in PATH` 时从 stdin 读入 armor 或二进制 age 密文。
- 未指定 `--out PATH` 时向 stdout 输出明文。
- 身份文件采用 age 标准解析：支持 age-keygen 风格文件，忽略空行与 `#` 注释行；同一文件可包含**多个身份**，任一匹配身份都能解密。
- 解密会先在权限受限的临时文件中完整校验认证标签；失败时不输出任何明文。
- 若 `--out` 已存在，会在**解密和创建临时明文之前**拒绝；已有文件绝不覆盖或删除。

stdin/stdout 示例：

```sh
safebox decrypt --identity PATH < secret.age
```

文件示例：

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

## 权限与不覆盖行为（摘要）

- `keygen`：私钥 `0600`，公钥 `0644`（POSIX 请求值，受 umask 限制）；任一文件都不覆盖。
- `encrypt`：输出文件 `0600`，绝不覆盖。
- `decrypt`：输出文件 `0600`，绝不覆盖；已存在的 `--out` 在产生任何明文之前就被拒绝；存在性检测也会拒绝悬空符号链接。
- 密文格式为 age 标准（默认 armor 文本，解密支持二进制）。

## 可以安全地问 AI 的问题示例

- “我在 `./keys/alice.txt` 有私钥，请给出把 `./report.pdf` 加密为 `./report.pdf.age`、收件人为 `AGE1...` 的确切 `safebox` 命令。”（只描述**路径**，绝不粘贴密钥）
- “解释 `--recipient` 和 `--identity` 的区别。”
- “起草一条把 `./secret.age` 解密到 stdout、不写文件的命令。”
- “`keygen` 请求什么权限？为什么这样是安全的？”

## 不要问 AI 的问题

- “这是我的私钥，帮我解密。”
- “这是我的明文/密文，看看里面是什么。”
- “帮我运行这条命令。”
- “打开 `./keys/alice.txt` 然后使用它。”

AI 应当拒绝这些请求，改为解释你自己应该运行的命令。
