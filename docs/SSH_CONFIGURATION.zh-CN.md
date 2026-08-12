# DistShip SSH 配置指南

DistShip 直接调用系统中的 `ssh` 和 `rsync`，不会读取、保存或复制密码与私钥。应先完成 SSH 配置并自行验证连接，再在 DistShip 中使用同一个 SSH 目标。

## DistShip 实际要求什么？

本地 SSH 配置不需要和下方示例完全一致。DistShip 只要求配置目标能够解析到正确服务器，并且所选用户可以通过公钥完成认证；部署用户还需要对远端目录具备写权限。

SSH 别名通常只需要 `HostName` 和 `User`，其他字段按实际情况使用：

- 服务器不使用 22 端口时才需要 `Port`。
- 默认密钥或 ssh-agent 已经能够选择正确身份时，可以不写 `IdentityFile`。
- 加载了多把密钥时推荐使用 `IdentitiesOnly yes`，但它不是所有连接的必填项。
- 保活参数可以提高长时间连接的稳定性，但不是 DistShip 的硬性要求。
- DistShip 不要求 `SetEnv`。服务器必须显式接受发送的环境变量，强制设置服务器未安装的区域设置还可能产生警告或失败。

DistShip 的预检查使用 `BatchMode=yes`，因此 `distship check` 期间不会弹出密码或密钥口令输入。带口令的密钥应提前加入 ssh-agent 或 macOS 钥匙串。

## 推荐配置

创建或编辑 `~/.ssh/config`：

```sshconfig
Host staging-web
    HostName 203.0.113.10
    User deploy
    Port 22
    IdentityFile ~/.ssh/id_ed25519_staging
    IdentitiesOnly yes
    ServerAliveInterval 30
    ServerAliveCountMax 3
```

将示例主机、用户和密钥路径替换为实际值。`Host` 后面的别名就是 DistShip 中使用的 SSH 目标。

收紧 SSH 文件权限：

```bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/config
chmod 600 ~/.ssh/id_ed25519_staging
chmod 644 ~/.ssh/id_ed25519_staging.pub
```

配置部署目标前先测试连接：

```bash
ssh staging-web
```

首次连接时，应将终端展示的服务器指纹与服务器管理员提供的可信值进行核对，确认一致后再接受。不要为了跳过提示而关闭主机密钥校验。

随后在初始化时使用该别名：

```text
SSH 目标（别名、主机名、IP 或 user@host）：staging-web
远端目录（绝对路径）：/var/www/example
```

对应的 TOML 配置为：

```toml
[projects.example.environments.test.target]
host = "staging-web"
directory = "/var/www/example"
```

首次部署前执行只读预检查：

```bash
distship check example:test
```

## 必须使用 SSH 别名吗？

不是。DistShip 支持 SSH 别名、域名、IP 和 `user@host`：

```text
staging-web
example.com
203.0.113.10
deploy@example.com
```

使用自定义端口、专用密钥、代理或跳板机时，推荐使用别名。将这些连接参数统一写入 `~/.ssh/config`，可以保证 `ssh` 与 `rsync` 使用完全一致的连接设置。

别名行为不符合预期时，可以查看 OpenSSH 最终采用的配置：

```bash
ssh -G staging-web | grep -E '^(hostname|user|port|identityfile|identitiesonly) '
```

`Host` 表示别名或匹配模式，`HostName` 才是实际连接目标。不要把某个服务器 IP 模式映射到无关的 `HostName`，否则使用该 IP 连接时会被重定向到配置中的其他目标。

## 配置顺序和作用域

第一个 `Host` 或 `Match` 之前的选项会全局生效，`# 旧服务器` 之类的注释不会形成作用域。OpenSSH 通常采用某个选项第一次取得的值，因此具体主机应放在前面，通用默认值放在最后：

```sshconfig
Host staging-web
    HostName example.com
    User deploy

Host *
    ServerAliveInterval 30
    ServerAliveCountMax 3
```

不要全局启用兼容算法或代理转发。

## GitHub 是独立的 SSH 目标

代码托管和部署服务器可以写在同一个配置文件中，但应使用独立的主机块。GitHub SSH 认证的固定远端用户是 `git`，不是 GitHub 账号邮箱：

```sshconfig
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/github_id_rsa
    IdentitiesOnly yes
```

它应与 DistShip 服务器连接分开验证：

```bash
ssh -T git@github.com
```

GitHub 根据账号中登记的公钥识别身份，不根据 `User` 字段选择账号。

## 自定义端口

端口应写入 SSH 别名，不写入 DistShip 项目配置：

```sshconfig
Host staging-web
    HostName example.com
    User deploy
    Port 10242
    IdentityFile ~/.ssh/id_ed25519_staging
    IdentitiesOnly yes
```

先使用 `ssh staging-web` 验证连接。DistShip 会在预检查和上传时复用同一个别名。

## 旧服务器的 `ssh-rsa` 兼容

使用 RSA 私钥不代表必须开启旧的 `ssh-rsa` 签名算法。现代 OpenSSH 可以让 RSA 密钥使用 SHA-2 签名。只有某个旧服务器明确报告找不到匹配的主机密钥或公钥签名算法时，才为该服务器单独增加兼容项：

```sshconfig
Host legacy-server
    HostName legacy.example.com
    User deploy
    HostKeyAlgorithms +ssh-rsa
    PubkeyAcceptedAlgorithms +ssh-rsa
```

不要把这两个选项放在第一个 `Host` 之前或 `Host *` 中，否则会为所有连接启用旧的 RSA/SHA-1 兼容。条件允许时应优先升级服务器。

## 跳板机

OpenSSH 可以通过堡垒机连接目标服务器，无需为 DistShip 增加专用字段：

```sshconfig
Host bastion
    HostName bastion.example.com
    User deploy
    IdentityFile ~/.ssh/id_ed25519_bastion
    IdentitiesOnly yes

Host staging-web
    HostName 10.0.0.20
    User deploy
    IdentityFile ~/.ssh/id_ed25519_staging
    IdentitiesOnly yes
    ProxyJump bastion
```

执行 `distship check` 前，先确认 `ssh staging-web` 可以正常连接。

## macOS 钥匙串

如果希望 macOS 记住密钥口令，可以在对应主机中增加：

```sshconfig
Host staging-web
    AddKeysToAgent yes
    UseKeychain yes
```

旧版或非 Apple OpenSSH 可能不识别 `UseKeychain`。需要跨系统共用配置时，可在文件顶部附近增加 `IgnoreUnknown UseKeychain`。

## SSH 代理转发

DistShip 不需要 `ForwardAgent yes`。只有某个可信服务器必须继续使用本机代理向其他主机认证时，才应在该主机块中单独启用。远端主机如果被攻破，攻击者虽然无法直接提取私钥，但可以在连接存续期间借助转发的代理执行认证操作。

## 常见问题

### Connection refused

表示已经访问到目标主机，但该端口没有 SSH 服务接受连接。检查 SSH 服务端口、防火墙、安全组、网络地址转换转发，以及所用别名中的 `Port`。

```bash
ssh -v staging-web
```

### Permission denied (publickey)

表示服务器接受了连接，但认证失败。检查远端用户、`IdentityFile`、密钥权限、服务器 `authorized_keys`，以及是否应设置 `IdentitiesOnly yes`。

```bash
ssh -v staging-web
ssh-add -l
```

如果 `ssh -i /密钥路径 user@host` 可以连接，而别名无法连接，应将相同的用户、主机、端口和密钥信息写入 SSH 别名，不要把私钥路径保存到 DistShip。

### 服务器主机密钥发生变化

先停止连接，确认服务器是否重装、SSH 主机密钥是否轮换，或者当前地址是否已经指向其他服务器。在通过独立渠道核对新指纹之前，不要直接删除 `known_hosts` 中的旧记录。

### SSH 正常但上传失败

确认本地和服务器都安装了 `rsync`，并确认 SSH 用户对目标目录具备写权限。`distship check <target-id>` 可以在不构建、不上传的情况下检查连接和远端目录状态。

## 安全规则

- 不要提交私钥、密码、令牌、`~/.ssh/config` 或个人部署配置。
- 不要在 `projects.toml` 中保存私钥路径，认证信息统一交给 SSH 配置管理。
- 优先使用只拥有目标目录权限的专用部署用户，不要默认使用无限制的 root 用户。
- 保持严格的主机密钥校验。
- DistShip 会以当前用户权限在本机执行构建命令，运行前应确认配置来源可信。
