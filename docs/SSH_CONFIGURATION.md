# SSH configuration for DistShip

DistShip uses the system `ssh` and `rsync` commands. It does not read, store, or copy passwords or private keys. Configure SSH first, verify the connection yourself, and then use the same SSH target in DistShip.

## What DistShip actually requires

Your SSH configuration does not need to match the example below. DistShip only requires that the configured target resolves to the intended server and that public-key authentication works for the selected user. The deployment user must also be able to write to the remote directory.

`HostName` and `User` are the usual minimum for an alias. The other fields are conditional:

- `Port` is only needed when the server does not use port 22.
- `IdentityFile` is optional when a default key or ssh-agent already selects the correct identity.
- `IdentitiesOnly yes` is recommended when several keys are loaded, but is not universally required.
- Keepalive options improve long-running connection reliability but are not required by DistShip.
- `SetEnv` is not required. The server must explicitly accept sent variables, and forcing a locale that is not installed remotely can produce warnings or failures.

DistShip preflight checks use `BatchMode=yes`. Password prompts and interactive key-passphrase prompts are therefore disabled during `distship check`. Add a passphrase-protected key to ssh-agent or the macOS keychain before running the check.

## Recommended setup

Create or edit `~/.ssh/config`:

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

Replace the example host, user, and key path with your own values. The alias after `Host` is the value used by DistShip.

Protect the SSH files:

```bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/config
chmod 600 ~/.ssh/id_ed25519_staging
chmod 644 ~/.ssh/id_ed25519_staging.pub
```

Test the connection before configuring a deployment target:

```bash
ssh staging-web
```

On the first connection, compare the displayed server fingerprint with a trusted value from the server administrator before accepting it. Do not disable host-key checking to suppress this prompt.

Then use the alias during initialization:

```text
SSH target (alias, host, IP, or user@host): staging-web
Remote directory (absolute path): /var/www/example
```

The resulting TOML target is:

```toml
[projects.example.environments.test.target]
host = "staging-web"
directory = "/var/www/example"
```

Run the read-only preflight check before the first deployment:

```bash
distship check example:test
```

## Is an SSH alias required?

No. DistShip accepts an SSH alias, hostname, IP address, or `user@host`:

```text
staging-web
example.com
203.0.113.10
deploy@example.com
```

An alias is recommended when the connection uses a custom port, dedicated key, proxy, jump host, or other SSH options. Keeping those options in `~/.ssh/config` ensures that `ssh` and `rsync` use the same connection settings.

Inspect the effective values selected by OpenSSH when an alias behaves unexpectedly:

```bash
ssh -G staging-web | grep -E '^(hostname|user|port|identityfile|identitiesonly) '
```

`Host` contains the alias or matching pattern; `HostName` contains the actual destination. Do not map a server IP pattern to an unrelated `HostName`, because connecting to that IP will be redirected to the configured destination.

## Ordering and scope

Options before the first `Host` or `Match` block apply globally. Comments such as `# legacy server` do not create a scope. OpenSSH normally uses the first value it obtains for an option, so put specific host blocks first and general defaults last:

```sshconfig
Host staging-web
    HostName example.com
    User deploy

Host *
    ServerAliveInterval 30
    ServerAliveCountMax 3
```

Avoid enabling compatibility algorithms or agent forwarding globally.

## GitHub is a separate SSH target

Git hosting and deployment servers can coexist in the same file, but they use separate host blocks. GitHub SSH authentication uses the fixed remote user `git`, not an account email address:

```sshconfig
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/github_id_rsa
    IdentitiesOnly yes
```

Test it independently from DistShip:

```bash
ssh -T git@github.com
```

The GitHub account is selected by the public key registered on GitHub, not by the `User` value.

## Custom port

Put the port in the SSH alias instead of the DistShip project configuration:

```sshconfig
Host staging-web
    HostName example.com
    User deploy
    Port 10242
    IdentityFile ~/.ssh/id_ed25519_staging
    IdentitiesOnly yes
```

Verify it with `ssh staging-web`. DistShip will reuse the same alias for preflight checks and uploads.

## Legacy `ssh-rsa` compatibility

An RSA private key does not automatically require the legacy `ssh-rsa` signature algorithm. Modern OpenSSH can use RSA keys with SHA-2 signatures. Add the legacy algorithm only when a specific old server reports that no matching host-key or public-key signature algorithm is available, and scope it to that host:

```sshconfig
Host legacy-server
    HostName legacy.example.com
    User deploy
    HostKeyAlgorithms +ssh-rsa
    PubkeyAcceptedAlgorithms +ssh-rsa
```

Do not place these directives before the first `Host` block or under `Host *`; that would enable legacy RSA/SHA-1 compatibility for every connection. Prefer upgrading the server when possible.

## Jump host

OpenSSH can route the deployment through a bastion without adding special DistShip fields:

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

Test `ssh staging-web` before running `distship check`.

## macOS keychain

To let macOS remember a key passphrase, add these options to the relevant host:

```sshconfig
Host staging-web
    AddKeysToAgent yes
    UseKeychain yes
```

Older or non-Apple OpenSSH builds may not recognize `UseKeychain`. If the same configuration is shared across systems, place `IgnoreUnknown UseKeychain` near the top of the file.

## Agent forwarding

`ForwardAgent yes` is not required by DistShip. Enable it only for a specific trusted host that must authenticate onward using your local agent. A compromised remote host can use the forwarded agent to perform authentication operations while the connection is active, even though it cannot directly extract the private key.

## Troubleshooting

### Connection refused

The host was reached, but nothing accepted the connection on that port. Check the SSH service port, firewall, security group, NAT forwarding, and the `Port` value in the selected alias.

```bash
ssh -v staging-web
```

### Permission denied (publickey)

The server accepted the connection but rejected authentication. Confirm the remote user, `IdentityFile`, key permissions, server `authorized_keys`, and whether `IdentitiesOnly yes` is appropriate.

```bash
ssh -v staging-web
ssh-add -l
```

If `ssh -i /path/to/key user@host` works but the alias does not, move the same user, host, port, and key into the alias configuration instead of storing the key path in DistShip.

### Host key changed

Stop and verify whether the server was rebuilt, its SSH host keys were rotated, or the address now points to a different machine. Do not delete the `known_hosts` entry until the new fingerprint has been independently verified.

### SSH works but upload fails

Confirm that `rsync` is installed locally and remotely, and that the SSH user can write to the configured directory. `distship check <target-id>` reports connection and remote-directory readiness without building or uploading.

## Security rules

- Never commit private keys, passwords, tokens, `~/.ssh/config`, or personal deployment configuration.
- Do not place a private-key path in `projects.toml`; keep authentication in SSH configuration.
- Prefer a dedicated deployment user with access limited to the required directory instead of unrestricted root access.
- Keep strict host-key verification enabled.
- Review configured build commands before running DistShip because they execute locally with your user permissions.
