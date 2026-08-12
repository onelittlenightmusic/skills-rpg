.PHONY: all build build-server build-mcp build-plugin cli-build cli-install cli-uninstall install-skills uninstall-skills install-claude uninstall-claude clean run-server stop-server restart-server restart reset smoke reload-go rebuild-frontend

BIN_DIR := bin
# Where the customs live in this repo, and where they get linked to. The
# targets below link rather than copy, so editing a custom here is editing the
# installed one — the two drifted apart back when this copied.
SKILL_SRC := server/skills
SKILL_DST ?= $(HOME)/.mywant/custom-types
DATA_DIR ?= $(HOME)/.mywant-rpg

# Where `make cli-install` puts the CLI — the same place mywant installs to.
# Override with: make cli-install INSTALL_DIR=/usr/local/bin
INSTALL_DIR ?= $(HOME)/.local/bin

all: build

# Always ensure full rebuild when YAML assets change
build: build-plugin smoke

build-plugin:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/mywant-rpg ./cmd/mywant-rpg  # Ensure CLI is built
	go build -o $(BIN_DIR)/stage-to-world ./cmd/stage-to-world

# Just the CLI — no stage-to-world, no smoke test. This is what cli-install
# needs, and it keeps bin/mywant-rpg current for the bin/rpg-server wrapper.
cli-build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/mywant-rpg ./cmd/mywant-rpg

# Build the CLI and put it on PATH.
cli-install: cli-build
	@mkdir -p $(INSTALL_DIR)
	@install -m 0755 $(BIN_DIR)/mywant-rpg $(INSTALL_DIR)/mywant-rpg
	@echo "installed: $(INSTALL_DIR)/mywant-rpg"
	@command -v mywant-rpg > /dev/null 2>&1 || \
	  echo "warning: $(INSTALL_DIR) is not on PATH"

cli-uninstall:
	@rm -f $(INSTALL_DIR)/mywant-rpg
	@echo "removed: $(INSTALL_DIR)/mywant-rpg"

smoke:
	./docs/smoke.sh

install-skills:
	@mkdir -p $(SKILL_DST)
	@for d in $(SKILL_SRC)/*/; do \
	  name=$$(basename $$d); \
	  src=$$(cd "$$d" && pwd); \
	  echo "linking $$name -> $(SKILL_DST)/$$name"; \
	  rm -rf "$(SKILL_DST)/$$name"; \
	  ln -s "$$src" "$(SKILL_DST)/$$name"; \
	done

uninstall-skills:
	@for d in $(SKILL_SRC)/*/; do \
	  name=$$(basename $$d); \
	  echo "removing $$name from $(SKILL_DST)"; \
	  rm -rf "$(SKILL_DST)/$$name"; \
	done

CLAUDE_SKILLS    := $(HOME)/.claude/skills
MYWANT_SKILLS    ?= $(HOME)/.mywant/custom-types/mywant-skills

install-claude:
	@mkdir -p $(CLAUDE_SKILLS)
	@for d in $(SKILL_SRC)/*/; do \
	  name=$$(basename $$d); \
	  echo "installing $$name -> $(CLAUDE_SKILLS)/$$name"; \
	  rm -rf "$(CLAUDE_SKILLS)/$$name"; \
	  cp -R "$$d" "$(CLAUDE_SKILLS)/$$name"; \
	done
	@if [ -d "$(MYWANT_SKILLS)" ]; then \
	  for d in $(MYWANT_SKILLS)/*/; do \
	    name=$$(basename $$d); \
	    echo "linking $$name -> $(CLAUDE_SKILLS)/$$name"; \
	    rm -rf "$(CLAUDE_SKILLS)/$$name"; \
	    ln -s "$$d" "$(CLAUDE_SKILLS)/$$name"; \
	  done; \
	else \
	  echo "warning: mywant-skills not found at $(MYWANT_SKILLS)"; \
	  echo "  see ~/.mywant/custom-types/mywant-skills/README.md for installation"; \
	fi

uninstall-claude:
	@for d in $(SKILL_SRC)/*/; do \
	  name=$$(basename $$d); \
	  rm -rf "$(CLAUDE_SKILLS)/$$name"; \
	  echo "removed $$name"; \
	done
	@if [ -d "$(MYWANT_SKILLS)" ]; then \
	  for d in $(MYWANT_SKILLS)/*/; do \
	    name=$$(basename $$d); \
	    rm -f "$(CLAUDE_SKILLS)/$$name"; \
	    echo "removed $$name"; \
	  done; \
	fi

clean:
	rm -rf $(BIN_DIR)

stop-server:
	@pkill -f 'bin/rpg-server' && echo "stopped" || echo "not running"

restart-server: stop-server
	@sleep 1
	$(MAKE) run-server

# build → background restart (no game state reset)
restart: build
	@pkill -f 'bin/rpg-server' || true
	@sleep 1
	@nohup $(BIN_DIR)/rpg-server > /tmp/rpg-server.log 2>&1 &
	@sleep 2
	@echo "rpg-server restarted (log: /tmp/rpg-server.log)"

# reset game state to initial (stages/*.yaml) without restarting server
reset:
	@curl -s -X POST http://localhost:7100/api/v1/reset > /dev/null && echo "game state reset"

smoke:
	bash docs/smoke.sh

# Rebuild frontend and restart mywant server to pick up the new bundle.
# Faster than restart-all because it skips the Go server rebuild.
MYWANT_DIR ?= $(HOME)/work/mywant
rebuild-frontend:
	@echo "building frontend..."
	@cd $(MYWANT_DIR)/web && npm run build
	@echo "rebuilding mywant CLI..."
	@cd $(MYWANT_DIR) && make build-cli
	@echo "restarting mywant server..."
	@curl -s -X POST http://localhost:8080/api/v1/system/restart > /dev/null
	@echo "done"

# Reload after Go server changes: rebuild → restart rpg-server → reload mywant.
reload-go: build-server
	@echo "restarting rpg-server..."
	@pkill -f 'bin/rpg-server' || true
	@sleep 1
	@nohup $(BIN_DIR)/rpg-server > /tmp/rpg-server.log 2>&1 &
	@sleep 2
	@echo "restarting mywant server..."
	@curl -s -X POST http://localhost:8080/api/v1/system/restart > /dev/null
	@echo "done — restart rpg_observe wants manually if needed"
