# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Cadence, please **do not open a public GitHub issue**. Instead, please report it responsibly by contacting us directly.

### How to Report

**Email**: [security@noslop.tech](mailto:security@noslop.tech)

**GitHub**: Open a [private security advisory](https://github.com/TryCadence/Cadence/security/advisories/new) on this repository (if available)

**Social Media**: [@NoSlopTech](https://x.com/NoSlopTech) on Twitter/X for urgent contact

### What to Include

When reporting a vulnerability, please include:

- Description of the vulnerability
- Steps to reproduce (if applicable)
- Potential impact
- Any proposed fixes (optional)

### Response Timeline

We will:
1. Acknowledge receipt of your report within 48 hours
2. Investigate and validate the issue
3. Work on a fix and release a patch
4. Credit you in the security advisory (unless you prefer anonymity)

### Security Updates

- Critical vulnerabilities will trigger immediate patch releases
- Non-critical security issues will be addressed in regular releases
- Security advisories will be published when fixes are available

## Supported Versions

| Version | Status | Support |
|---------|--------|---------|
| v0.3.x  | Current | Full support |
| v0.2.x  | Legacy | Security fixes only |
| < v0.1  | EOL | Not supported |

## Security Best Practices

When using Cadence:

- Keep your installation updated to the latest version
- Review analysis output carefully before taking action
- Use configuration files securely (don't commit secrets)
- Run in isolated environments when analyzing untrusted repositories

## Dependencies

Cadence uses the following key dependencies:

- `go-git/v5` - Git operations
- `PuerkitoBio/goquery` - HTML parsing
- `spf13/cobra` - CLI framework
- `openai-go` - Optional AI analysis

We monitor these dependencies for security issues and update promptly when vulnerabilities are found.

## Questions?

For security-related questions (not vulnerability reports), please contact [security@noslop.tech](mailto:security@noslop.tech).
