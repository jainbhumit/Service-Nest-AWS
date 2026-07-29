# Service Nest — AWS Backend

Go Lambda API (Gorilla Mux) behind API Gateway. Single production stack in `us-east-1`.

**Prod API (current):** `https://1fh0244il4.execute-api.us-east-1.amazonaws.com/Prod`

## Prerequisites

- Go 1.21+
- [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (local DynamoDB only)
- AWS CLI configured for manual deploys (`AWS_PROFILE`)

## Local development

```powershell
cd Service-Nest-AWS
copy .env.example .env
# Edit JWT_SECRET and other values in .env

make local
# or: powershell -File scripts/run-local.ps1
```

| Endpoint | Purpose |
|----------|---------|
| `http://localhost:8080/health` | Liveness |
| `http://localhost:8080/health/ready` | DynamoDB connectivity |

Local stack uses DynamoDB Local (port 8000). SNS is disabled when `SNS_TOPIC_ARN` is empty. S3 and SMTP are optional locally.

### Makefile targets

| Target | Description |
|--------|-------------|
| `make test` | `go test ./...` |
| `make validate` | Tests + `sam validate` |
| `make local` | Compose + table init + local HTTP server |
| `make deploy` | Manual prod deploy (interactive changeset) |
| `make verify-deploy` | Post-deploy smoke bundle |

## Manual production deploy

Use when GitHub Actions is unavailable or for emergency deploys from your machine.

```powershell
cd Service-Nest-AWS
copy .env.example .env
# Set JWT_SECRET, SMTP_APP_PASSWORD, SMTP_FROM

$env:AWS_PROFILE = "your-prod-profile"
make deploy
make verify-deploy
```

Stack: **`serviceNest`** · Region: **`us-east-1`**

Secrets are passed as SAM parameters (`JwtSecret`, `SmtpAppPassword`, `SmtpFrom`) — never committed to git.

## GitHub OIDC setup (one-time)

GitHub Actions deploys use **OIDC only** — no long-lived `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` in secrets.

### 1. GitHub OIDC provider in AWS

If not already present in account `116777895904`:

- Provider URL: `https://token.actions.githubusercontent.com`
- Audience: `sts.amazonaws.com`

### 2. IAM role

Create role **`GitHubActionsServiceNestDeploy`** with:

- **Trust policy:** [`docs/github-oidc-trust-policy.json`](docs/github-oidc-trust-policy.json)  
  Restricts assumption to `repo:jainbhumit/Service-Nest-AWS:environment:production` (fork PRs cannot assume this role).

- **Permissions policy:** [`docs/github-oidc-permissions-policy.json`](docs/github-oidc-permissions-policy.json)  
  Scoped to stack `serviceNest` and SAM artifact buckets. Adjust if the first deploy surfaces missing actions.

### 3. GitHub repository settings

**Environment `production`** (Settings → Environments):

- Enable **Required reviewers** (self-approval is fine for a solo repo).
- Add environment secrets:

| Secret | Description |
|--------|-------------|
| `AWS_DEPLOY_ROLE_ARN` | ARN of `GitHubActionsServiceNestDeploy` |
| `JWT_SECRET` | JWT signing secret (16+ chars) |
| `SMTP_APP_PASSWORD` | Gmail app password for OTP |
| `SMTP_FROM` | Sender email address |

Do **not** store AWS access keys. OTP values must never appear in workflow logs.

## CI — pull requests

Workflow: [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

On every PR to `main`:

1. `go test ./...`
2. `sam validate`

No AWS credentials. No deploy.

## Production deploy — approval gate

Workflow: [`.github/workflows/deploy-prod.yml`](.github/workflows/deploy-prod.yml)

**Triggers:**

- Push to `main` (after merge)
- Manual `workflow_dispatch`

**Flow:**

```
test job (go test + sam validate)
    ↓
deploy job → waits for production environment approval
    ↓
OIDC → sam build → sam deploy (stack serviceNest)
    ↓
scripts/verify-deploy.sh (smoke bundle)
```

The deploy job logs the deployed SHA as **last known-good candidate** — note it after successful runs.

If smoke fails after a successful CloudFormation update, treat prod as unhealthy and roll back before debugging live.

## Rollback

| Method | When to use |
|--------|-------------|
| **GitHub Actions** | Preferred — Actions → Deploy Production → Run workflow → set `git_ref` to a previous commit SHA or tag → approve deploy |
| **Local** | `git checkout <good-sha>` → `make deploy` → `make verify-deploy` |
| **CloudFormation auto-rollback** | If `sam deploy` fails mid-update, CloudFormation rolls back the changeset automatically |

After rollback, re-run the smoke bundle (`make verify-deploy` or wait for the workflow smoke step).

## Post-deploy checklist (prod-only)

These cannot be fully verified locally:

1. **Smoke bundle** — `make verify-deploy` or CI smoke step (health, ready, login error envelope).
2. **Category image upload** — presigned S3 PUT from the Angular UI.
3. **OTP email** — trigger login/signup with a real address; confirm email delivery.
4. **SNS alerts** (optional) — induce a 500 in a controlled test; confirm SNS notification.

## Troubleshooting

| Symptom | Check |
|---------|-------|
| `GET /health/ready` fails after deploy | Lambda IAM → DynamoDB table `servicenest`; region `us-east-1` |
| Login works but no OTP email | `SMTP_FROM`, `SMTP_APP_PASSWORD` in Lambda env / GitHub secrets |
| OIDC `Not authorized to perform sts:AssumeRoleWithWebIdentity` | Trust policy `sub` matches `environment:production`; role ARN secret correct |
| SAM deploy permission denied | Extend [`docs/github-oidc-permissions-policy.json`](docs/github-oidc-permissions-policy.json) |
| Local DynamoDB errors | Docker running; `make local-init` or `scripts/local-init.ps1` |

## Project layout

```
Service-Nest-AWS/
├── .github/workflows/     # ci.yml, deploy-prod.yml
├── docs/                  # OIDC IAM JSON templates
├── scripts/               # deploy, verify-deploy, local-init
├── service-nest/          # Go source (cmd/main.go = Lambda, cmd/local = dev server)
├── template.yaml          # SAM template
└── samconfig.toml         # Stack serviceNest, us-east-1
```
