# Ismail Ahmed: The comprehensive and modern cross-platform Makefile for the Python `obliv` package

# Initialzing all used variables
VENV_SCRIPT=venv.sh
VENV_DIR=venv
.PHONY: all setup test clean

# Always building first using the venv script
all: setup

setup:
	@echo "[MAKE] Setting up environment..."
	@chmod +x $(VENV_SCRIPT)
	@bash $(VENV_SCRIPT) $(VENV_DIR)

# Running the tox tests
test: setup
	@echo "[MAKE] Running tests with tox..."
	@source $(VENV_DIR)/bin/activate && tox

# Cleaning up all post-run items
clean:
	@echo "[MAKE] Cleaning up virtual environment and tox artifacts..."
	rm -rf $(VENV_DIR) .tox .pytest_cache .mypy_cache dist build *.egg-info
