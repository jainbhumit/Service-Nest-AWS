.PHONY: build test validate deploy deploy-guided verify-deploy compose-up local-init run-local local

# Prefer bash scripts on Linux/WSL; PowerShell on native Windows.
ifeq ($(OS),Windows_NT)
  LOCAL_INIT_CMD = powershell -ExecutionPolicy Bypass -File scripts/local-init.ps1
  RUN_LOCAL_CMD = powershell -ExecutionPolicy Bypass -File scripts/run-local.ps1
  DEPLOY_CMD = powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1
  DEPLOY_GUIDED_CMD = powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1 -Guided
  VERIFY_CMD = powershell -ExecutionPolicy Bypass -File scripts/verify-deploy.ps1
else
  LOCAL_INIT_CMD = bash scripts/local-init.sh
  RUN_LOCAL_CMD = bash scripts/run-local.sh
  DEPLOY_CMD = bash scripts/deploy.sh
  DEPLOY_GUIDED_CMD = bash scripts/deploy.sh --guided
  VERIFY_CMD = bash scripts/verify-deploy.sh
endif

build:
	sam build

test:
	cd service-nest && go test ./...

validate: test
	sam validate

deploy:
	$(DEPLOY_CMD)

deploy-guided:
	$(DEPLOY_GUIDED_CMD)

verify-deploy:
	$(VERIFY_CMD)

compose-up:
	docker compose up -d

local-init:
	$(LOCAL_INIT_CMD)

run-local:
	cd service-nest && go run ./cmd/local

local:
	$(RUN_LOCAL_CMD)
