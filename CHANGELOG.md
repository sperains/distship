# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project follows Semantic Versioning.

## Unreleased

- Initial project skeleton.
- Interactive configuration initialization.
- Configuration validation and grouped project listing.
- Built-in English and Simplified Chinese terminal interfaces with automatic locale detection.
- Streamlined initialization with detected defaults, an advanced mode, and automatic post-write validation.
- Accepted project directories as command arguments or required interactive input, with immediate validation and explicit detection results.
- Added explicit current-project detection and flexible SSH targets using aliases, hostnames, IP addresses, or user-qualified hosts.
- Accepted deployment targets as either a combined SSH target and remote path or two guided inputs.
- Replaced duplicate overwrite prompts with field-level diffs and a single confirmation for existing targets.
- Safe local target removal with confirmation, validation, and timestamped backup of the last configuration.
- Standardized target selection on the copyable `project:environment` ID.
- Reused configured target IDs when initialization is run again for the same local directory.
