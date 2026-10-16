# safebox

`safebox` 是一个安全、可跨平台的 Go 命令行加密工具，用于加密本地文件与文本，也适合在应用之间通过管道传输密文。

- 只使用成熟的 [age](https://age-encryption.org) 标准库（`filippo.io/age`），**不自定义任何加密算法或协议**。
- 密文默认使用 age 的 **armor 文本格式**（`-----BEGIN AGE ENCRYPTED FILE-----`），方便复制粘贴和管道传输。
- 默认读写 stdin/stdout，支持 shell 与 PowerShell 管道；**明文不能作为命令行参数传入**，避免泄露到 shell 历史和进程列表。
- 这是一个纯本地 CLI，**不会上传任何数据**。

## 安装与构建

需要 Go 1.25 或更高版本（与当前 `filippo.io/age` 依赖的最低版本一致）。

```sh
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox
go build -o safebox .
```

跨平台交叉编译（纯 Go，无需 CGO）：

```sh
# Linux amd64
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
# Windows
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

安装到 `$GOBIN` / `%GOPATH%\bin`：

```sh
go install .
```

### 预编译版本

`dist/` 目录提供 Linux、macOS 与 Windows 的 amd64、arm64 版本；请按操作系统和 CPU 架构选择对应文件。`dist/SHA256SUMS` 用于检查下载是否损坏，但校验和本身不证明发布者身份。当前二进制尚未进行代码签名；高敏感用途建议从源码自行构建并检查来源。

查看版本与帮助：

```sh
safebox version
safebox help
```

## 命令

### 生成密钥对

```sh
safebox keygen --identity ./keys/alice.txt
```

- 私钥写入 `./keys/alice.txt`，向 POSIX 系统请求权限 `0600`；公钥写入 `./keys/alice.txt.pub`，请求权限 `0644`。
- POSIX 系统上的**实际权限还会受进程 umask 限制**：umask 只会移除权限位，因此实际权限不会比请求值更宽。例如 umask `022` 下公钥为 `0644`，umask `077` 下为 `0600`；私钥请求的 `0600` 不含组/其他位，通常不受 umask 影响。
- 在 Windows 上，文件访问权限由父目录继承的 ACL 控制；请确认父目录仅授权给预期账户。
- 若任一文件已存在则拒绝，**绝不覆盖**。
- 公钥（收件人）会同时打印到 stdout，便于分发。

### 加密

默认从 stdin 读入明文，向 stdout 输出 armor 密文：

```sh
safebox encrypt --recipient age1... < plain.txt > secret.age
```

可以重复指定多个收件人（任意一个私钥都能解密）：

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb < plain.txt > secret.age
```

也可以使用文件路径（输出文件权限 `0600`，且**拒绝覆盖**已有文件）：

```sh
safebox encrypt --recipient age1... --in plain.txt --out secret.age
```

POSIX 系统输出文件权限请求为 `0600`（同样受 umask 限制，只会更窄、不会更宽）；Windows 使用父目录继承的 ACL。

### 解密

默认从 stdin 读入密文，向 stdout 输出明文：

```sh
safebox decrypt --identity ./keys/alice.txt < secret.age > plain.txt
```

文件方式（POSIX 系统输出权限 `0600`，Windows 使用继承的 ACL，且**拒绝覆盖**）：

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

身份文件采用 age 标准解析：支持 age-keygen 风格的文件，忽略空行与 `#` 注释行（例如 `# created: ...`、`# public key: age1...`），并允许同一文件内包含**多个身份**，任一匹配身份都能解密。

解密会先将明文流式暂存到系统临时目录中权限受限的临时文件，完整校验认证标签后才写出；认证失败时不会输出任何明文，也不会因文件过大而占用同等规模的内存。临时文件在退出时删除，但删除**不等于安全擦除**；临时目录所在磁盘、交换分区或备份仍可能留下数据。输入同时支持 armor 文本和二进制 age 格式。

若 `--out` 指向已存在的路径，会在**解密和创建临时明文之前**就拒绝；该存在性检测不跟随符号链接，因此悬空符号链接同样会被视为已存在而拒绝，最终写出仍使用 `O_EXCL` 防止竞态。已有的 `--out` 目标在任何情况下都不会被覆盖或删除；但若本次操作已开始创建新文件后遇到写入失败，会清理掉本次产生的部分输出。

## 完整示例

### 文本管道（Linux / macOS / Git Bash）

```sh
printf 'hello safebox\n' | safebox encrypt --recipient "$(cat alice.txt.pub)" > msg.age
safebox decrypt --identity alice.txt < msg.age
```

### 文件

```sh
safebox keygen --identity alice.txt
safebox encrypt --recipient "$(cat alice.txt.pub)" --in report.pdf --out report.pdf.age
safebox decrypt --identity alice.txt --in report.pdf.age --out report.pdf
```

### 应用之间传输（管道）

```sh
# 应用 A 生成密文写入管道
producer-app | safebox encrypt --recipient age1... | consumer-app
```

### Windows PowerShell

```powershell
# 生成密钥
safebox keygen --identity .\keys\alice.txt

# 文件方式避免 PowerShell 文本管道改写编码/换行
safebox encrypt --recipient (Get-Content -Raw .\keys\alice.txt.pub).Trim() `
    --in .\plain.txt --out .\secret.age

# 解密
safebox decrypt --identity .\keys\alice.txt `
    --in .\secret.age --out .\plain.out.txt
```

> 注意：文件模式不会经过 PowerShell 的文本对象管道，适合任意文件。若直接在终端输入敏感文本作为 `echo`/`printf` 命令内容，shell 历史仍可能记录它；敏感文本请从文件、剪贴板工具或上游应用经 stdin 传入。不要把明文或私钥拼进 `safebox` 命令行。

## 安全边界

- **加密算法**：使用 age / X25519 标准协议，密钥与密文格式由 `filippo.io/age` 负责，本工具不发明密码学。
- **私钥保护**：POSIX 系统私钥文件以 `0600` 创建；Windows 由父目录 ACL 控制。`safebox` 绝不打印私钥。请自行妥善保护私钥文件，不要在多人共用或未加密的磁盘上存放。
- **临时明文**：解密时会使用系统临时目录下权限受限的临时文件完成整流校验；用完后删除，但不保证从 SSD、交换分区、备份或快照中安全擦除。需要高保密时，请确保临时目录与交换空间位于可信的加密存储上。
- **不可恢复**：私钥一旦丢失，密文**无法恢复**；没有后门、没有找回机制。
- **元数据**：加密只保护文件内容，age 不隐藏文件大小、时间、文件名等元数据，也不隐藏流量特征。
- **终端与剪贴板**：通过终端复制粘贴密钥或明文时，内容可能残留在终端历史、回滚缓冲或剪贴板中。
- **口令熵**：本工具使用公钥加密，不依赖口令；如需口令加密请勿使用低熵口令。
- **本地工具**：`safebox` 是一个本地 CLI，不联网、不上传数据、不提供服务端功能。

## 开发与测试

```sh
gofmt -l .
go vet ./...
go test ./...
```

测试覆盖密钥生成与文件权限（按平台适配）、stdin/stdout 与文件往返、多收件人、密文被篡改或私钥错误时失败且不泄露明文、以及拒绝覆盖已有文件。

## 许可证

MIT，见 [LICENSE](LICENSE)。
