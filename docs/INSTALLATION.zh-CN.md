# 安装 DistShip

DistShip 为 macOS 和 Linux 提供带校验和的预构建压缩包。可以通过 Homebrew、
Codex 技能或正式版本压缩包安装。

## 支持平台

- macOS：支持 Apple Silicon 和 Intel，压缩包名称包含 `Darwin`。
- Linux：支持 ARM64 和 x86-64，压缩包名称包含 `Linux`。

当前暂不支持 Windows。

## 使用 Homebrew 安装

```bash
brew install --cask sperains/tap/distship
```

Homebrew 会安装与当前 macOS 或 Linux 平台匹配的二进制，并根据 Cask 中记录的
校验和验证压缩包。

## 使用 Codex 安装

让 Codex 从本仓库安装配套技能：

```text
从 https://github.com/sperains/distship/tree/main/skills/install-distship 安装技能，
然后使用 $install-distship 安装最新稳定版本。
```

技能会识别操作系统和架构，下载对应版本，验证 SHA-256 校验和，并默认安装到
`$HOME/.local/bin`。也可以指定版本或安装目录。

## 手动安装

### 1. 下载

从 [GitHub 最新版本](https://github.com/sperains/distship/releases/latest)
下载与当前平台匹配的压缩包和 `checksums.txt`，并将两个文件放在同一目录。

### 2. 校验压缩包

在下载目录中执行当前操作系统对应的命令：

```bash
# macOS
shasum -a 256 --ignore-missing --check checksums.txt

# Linux
sha256sum --ignore-missing --check checksums.txt
```

只有下载的压缩包显示 `OK` 时才继续安装。

### 3. 安装二进制

将 `<压缩包>` 替换为实际下载的文件名：

```bash
tar -xzf <压缩包>
mkdir -p "$HOME/.local/bin"
install -m 0755 distship "$HOME/.local/bin/distship"
"$HOME/.local/bin/distship" version
```

如果无法直接执行 `distship`，将下面一行加入终端配置文件，再打开新终端：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 升级

通过 Homebrew 安装时，执行：

```bash
brew upgrade --cask distship
```

其他安装方式可以再次运行 Codex 技能，或下载并校验新版本后替换现有二进制。
升级不会修改 DistShip 配置和本地部署历史。

升级后确认当前生效版本：

```bash
distship version
```

## 卸载

通过 Homebrew 安装时，执行：

```bash
brew uninstall --cask distship
```

执行 `command -v distship` 找到当前生效的二进制，再删除该文件。配置和部署历史会默认保留。

只有确定不再需要本地记录时，才清理以下目录：

```text
${XDG_CONFIG_HOME:-$HOME/.config}/distship
${XDG_STATE_HOME:-$HOME/.local/state}/distship
```

安装完成后，返回[快速开始](../README.zh-CN.md#快速开始)继续配置。
