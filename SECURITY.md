# Security Policy

DistShip does not store passwords or private keys. It reuses the user's SSH configuration and invokes local Git, SSH, rsync, and project build tools.

Only load configuration files you trust. Configured build commands can execute programs on the local machine.

Do not report security vulnerabilities in public issues. Until a private reporting address is published, contact the repository owner through a private GitHub channel.
