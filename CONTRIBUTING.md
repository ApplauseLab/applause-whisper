# Contributing to Yap

## Git Workflow

### Branch Naming Convention

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feat/<description>` | `feat/customizable-hotkeys` |
| Bug fix | `fix/<description>` | `fix/audio-recording-crash` |
| Docs | `docs/<description>` | `docs/update-readme` |
| Refactor | `refactor/<description>` | `refactor/simplify-config` |

### Development Flow

1. **Create feature branch**
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feat/your-feature-name
   ```

2. **Make commits using Conventional Commits**
   ```bash
   git commit -m "feat: add new feature description"
   git commit -m "fix: resolve bug description"
   ```

3. **Push and create PR**
   ```bash
   git push origin feat/your-feature-name
   gh pr create --title "feat: your feature" --body "Description of changes"
   ```

4. **Merge PR to main** (after review)

5. **Release Please auto-creates release PR**
   - Analyzes commits since last release
   - Updates CHANGELOG.md
   - Bumps version numbers

6. **Merge release PR when ready to publish**
   - Creates git tag (e.g., v0.3.0)
   - Triggers build workflow
   - Publishes GitHub release with binaries

## Conventional Commits

We use [Conventional Commits](https://www.conventionalcommits.org/) for automatic versioning and changelog generation.

### Commit Message Format

```
<type>: <description>

[optional body]

[optional footer]
```

### Types and Version Bumps

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat:` | New feature | MINOR (0.2.0 -> 0.3.0) |
| `fix:` | Bug fix | PATCH (0.2.0 -> 0.2.1) |
| `docs:` | Documentation only | None |
| `style:` | Code style (formatting) | None |
| `refactor:` | Code refactoring | None |
| `perf:` | Performance improvement | PATCH |
| `test:` | Adding tests | None |
| `chore:` | Maintenance tasks | None |
| `ci:` | CI/CD changes | None |
| `build:` | Build system changes | None |

### Breaking Changes

For breaking changes, add `!` after the type or include `BREAKING CHANGE:` in the footer:

```bash
git commit -m "feat!: redesign configuration API"
# or
git commit -m "feat: redesign configuration API

BREAKING CHANGE: config.json schema has changed"
```

This triggers a MAJOR version bump (0.2.0 -> 1.0.0).

## Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH (e.g., 0.3.1)
  |     |     |
  |     |     +-- PATCH: Bug fixes, no new features
  |     +-------- MINOR: New features, backwards compatible
  +-------------- MAJOR: Breaking changes
```

## Release Process

1. **Automatic**: Release Please monitors `main` branch
2. **On new commits**: Creates/updates a release PR with changelog
3. **When ready**: Merge the release PR
4. **Automated**: Tag created -> Build runs -> Release published

### Manual Release (if needed)

```bash
git tag v0.3.0
git push origin v0.3.0
```

This triggers the build workflow directly.

## Files Updated by Release Please

- `CHANGELOG.md` - Release notes
- `wails.json` - `info.productVersion`
- `package.json` - `version`
- `frontend/package.json` - `version`
- `.release-please-manifest.json` - Internal tracking
