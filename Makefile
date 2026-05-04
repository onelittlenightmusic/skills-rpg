.PHONY: all build build-server build-mcp install-skills uninstall-skills install-claude uninstall-claude clean run-server stop-server restart-server smoke

BIN_DIR := bin
SKILL_DST ?= $(HOME)/.mywant/custom-types
DATA_DIR ?= $(HOME)/.mywant-rpg

all: build

build: build-server build-mcp

build-server:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/rpg-server ./cmd/rpg-server

build-mcp:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/rpg-mcp ./cmd/rpg-mcp

run-server: build-server
	$(BIN_DIR)/rpg-server

install-skills:
	@mkdir -p $(SKILL_DST)
	@for d in skills/*/; do \
	  name=$$(basename $$d); \
	  echo "installing $$name -> $(SKILL_DST)/$$name"; \
	  rm -rf "$(SKILL_DST)/$$name"; \
	  cp -R "$$d" "$(SKILL_DST)/$$name"; \
	done
	@for f in generated-want-types/*.yaml; do \
	  [ -e "$$f" ] || continue; \
	  echo "installing $$(basename $$f)"; \
	  cp "$$f" "$(SKILL_DST)/"; \
	done

uninstall-skills:
	@for d in skills/*/; do \
	  name=$$(basename $$d); \
	  echo "removing $$name from $(SKILL_DST)"; \
	  rm -rf "$(SKILL_DST)/$$name"; \
	done
	@for f in generated-want-types/*.yaml; do \
	  [ -e "$$f" ] || continue; \
	  rm -f "$(SKILL_DST)/$$(basename $$f)"; \
	  echo "removed $$(basename $$f)"; \
	done

CLAUDE_SKILLS    := $(HOME)/.claude/skills
MYWANT_SKILLS    ?= $(HOME)/.mywant/custom-types/mywant-skills

install-claude:
	@mkdir -p $(CLAUDE_SKILLS)
	@for d in skills/*/; do \
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
	@for d in skills/*/; do \
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

smoke:
	bash docs/smoke.sh
