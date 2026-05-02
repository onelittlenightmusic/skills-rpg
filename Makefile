.PHONY: all build build-server build-mcp install-skills clean run-server smoke

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

clean:
	rm -rf $(BIN_DIR)

smoke:
	bash docs/smoke.sh
