---
name: go-featuremanagement-release
description: "Release the Go FeatureManagement package and its feature flag provider. Use when: releasing a new version, bumping version, tagging a release, publishing to Go proxy, updating provider dependency, creating release PRs."
argument-hint: "Target version number, e.g. 1.2.0"
---

# Go FeatureManagement Release

## When to Use

- Release a new version of the Go FeatureManagement package (`featuremanagement`) and its feature flag provider (`featuremanagement/providers/azappconfig`)
- Bump the module version and publish to the Go module proxy

## Prerequisites

- GitHub CLI (`gh`) ≥ 2.86.0
- Authenticated via `gh auth login` with `repo` and `read:org` scopes
- GitHub Copilot agent enabled for the repository
- Write access to `microsoft/FeatureManagement-Go`

## Cycle: Create → Monitor Task → Monitor PR

Every PR in this release follows the same three-phase cycle. The steps below reference this cycle.

### A. Create the agent task

Create one agent task for the next PR in the sequence. Each command returns a URL in the format:

`https://github.com/microsoft/FeatureManagement-Go/pull/<pr-id>/agent-sessions/<agent-session-id>`

### B. Monitor the agent task

Extract the `agent-session-id` from the URL returned in step A, then poll:

```bash
gh agent-task view <agent-session-id>
```

Keep polling until the session state is `Ready for review`. Print the PR URL for reference.

### C. Monitor the PR until merged

Poll the PR every 10 minutes. Stop monitoring if:

- PR state is `MERGED` → proceed to the next step (or finish if this was the last PR).
- PR state is `CLOSED` or `Abandoned` → report and **stop the release**.
- 24 hours elapsed → report current status and **stop the release**.

```bash
gh pr view <pr-id> --repo microsoft/FeatureManagement-Go --json state --jq '.state'
```

---

## Procedure

Follow these steps **in order**. Each step depends on the previous one completing successfully.

### Step 1 — Create Version Bump PR

Use the **Create → Monitor Task → Monitor PR** cycle:

**A. Create the agent task:**

```bash
gh agent-task create \
  "Please create a version bump PR for version <version>. \
  Checkout a new branch from release/v<version> (e.g. version-bump/v<version>), \
  update the moduleVersion in featuremanagement/version.go to <version>, \
  and open a PR targeting the release/v<version> branch with title 'Version bump v<version>'." \
  --repo microsoft/FeatureManagement-Go \
  --base release/v<version>
```

**B. Monitor the agent task** until session state is `Ready for review`.

**C. Monitor the PR** until it is merged. Do not proceed until the PR state is `MERGED`.

### Step 2 — Tag the Release

After the version bump PR is merged, fetch the release branch and create a git tag at the HEAD:

```bash
git fetch origin release/v<version>
git tag featuremanagement/v<version> origin/release/v<version>
```

### Step 3 — Push the Tag

Push the tag to the remote:

```bash
git push origin featuremanagement/v<version>
```

### Step 4 — Publish FeatureManagement to Go Module Proxy

**Before executing the publish command**, generate a summary report table for human review:

| Item                | Detail                                                              |
|---------------------|---------------------------------------------------------------------|
| **Version**         | `v<version>`                                                        |
| **Version file**    | `featuremanagement/version.go` updated to `<version>`              |
| **Tag pushed**      | `featuremanagement/v<version>`                                      |
| **Publish command** | `GOPROXY=proxy.golang.org go list -m github.com/microsoft/Featuremanagement-Go/featuremanagement@v<version>` |
| **Next step**       | After publish, create feature flag provider dependency update PR    |

> **Pause here.** Present the table and wait for the user to confirm before proceeding.

After user confirmation, run:

```bash
GOPROXY=proxy.golang.org go list -m github.com/microsoft/Featuremanagement-Go/featuremanagement@v<version>
```

### Step 5 — Create Feature Flag Provider Dependency Update PR

Use the **Create → Monitor Task → Monitor PR** cycle:

**A. Create the agent task:**

```bash
gh agent-task create \
  "Please create a PR to update the feature flag provider dependency. \
  Checkout a new branch from release/v<version> (e.g. update-provider-dep/v<version>). \
  In featuremanagement/providers/azappconfig/go.mod, update the featuremanagement dependency version: \
  change 'require github.com/microsoft/Featuremanagement-Go/featuremanagement v<old-version>' \
  to 'require github.com/microsoft/Featuremanagement-Go/featuremanagement v<version>'. \
  Then run 'cd featuremanagement/providers/azappconfig && go mod tidy' to update go.sum. \
  Open a PR targeting the release/v<version> branch with title 'Feature flag provider: update dependency v<version>'." \
  --repo microsoft/FeatureManagement-Go \
  --base release/v<version>
```

**B. Monitor the agent task** until session state is `Ready for review`.

**C. Monitor the PR** until it is merged. Do not proceed until the PR state is `MERGED`.

### Step 6 — Tag the Feature Flag Provider Release

After the dependency update PR is merged, fetch the release branch and create a git tag for the provider at the HEAD:

```bash
git fetch origin release/v<version>
git tag featuremanagement/providers/azappconfig/v<version> origin/release/v<version>
```

### Step 7 — Push the Provider Tag

Push the provider tag to the remote:

```bash
git push origin featuremanagement/providers/azappconfig/v<version>
```

### Step 8 — Publish Feature Flag Provider to Go Module Proxy

**Before executing the publish command**, generate a summary report table for human review:

| Item                | Detail                                                              |
|---------------------|---------------------------------------------------------------------|
| **Version**         | `v<version>`                                                        |
| **Provider module** | `featuremanagement/providers/azappconfig`                           |
| **Dependency**      | `featuremanagement v<version>`                                      |
| **Tag pushed**      | `featuremanagement/providers/azappconfig/v<version>`                |
| **Publish command** | `GOPROXY=proxy.golang.org go list -m github.com/microsoft/Featuremanagement-Go/featuremanagement/providers/azappconfig@v<version>` |
| **Next step**       | After publish, create merge-back PR (release branch → main)        |

> **Pause here.** Present the table and wait for the user to confirm before proceeding.

After user confirmation, run:

```bash
GOPROXY=proxy.golang.org go list -m github.com/microsoft/Featuremanagement-Go/featuremanagement/providers/azappconfig@v<version>
```

### Step 9 — Create Merge-Back PR

Use the **Create → Monitor Task → Monitor PR** cycle:

**A. Create the agent task:**

```bash
gh agent-task create \
  "Please create a PR to merge release/v<version> back to main with title 'Merge release/v<version> to main'." \
  --repo microsoft/FeatureManagement-Go \
  --base main
```

**B. Monitor the agent task** until session state is `Ready for review`.

**C. Monitor the PR** until it is merged. The release is complete once this PR is merged.

## Notes

- The version in `featuremanagement/version.go` uses the format `X.Y.Z` (no `v` prefix) in the `moduleVersion` constant.
- The featuremanagement module tag uses the format `featuremanagement/vX.Y.Z` (with `v` prefix and module path prefix).
- The feature flag provider tag uses the format `featuremanagement/providers/azappconfig/vX.Y.Z`.
- The publish commands only change the version portion: `@vX.Y.Z`.
- The feature flag provider dependency update PR modifies `featuremanagement/providers/azappconfig/go.mod` and `go.sum` (via `go mod tidy`).
