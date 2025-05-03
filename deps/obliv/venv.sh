#!/usr/bin/env bash
# Ismail Ahmed: An automated Python 3.12 virtual enviroment script for cross-platform building and testing


# # Setting various variables to their default states
# set -euo pipefail
# VENV_NAME=${1:-venv}
# REQUIRED_VERSION="3.12"
# FOUND_PY=""

# # Trying to find common python3.12 binary candidate names and paths
# echo "[INFO] Looking for Python $REQUIRED_VERSION..."
# CANDIDATES=(
#   python3.12
#   /usr/bin/python3.12
#   /usr/local/bin/python3.12
#   "$HOME/.pyenv/versions/3.12.*/bin/python"
#   "$HOME/.local/bin/python3.12"
# )

# # Iterating over the candidates to find a working one
# for CAND in "${CANDIDATES[@]}"; do
#   if command -v $CAND &>/dev/null; then
#     VERSION=$($CAND -c 'import sys; print(".".join(map(str, sys.version_info[:2])))')
#     if [[ "$VERSION" == "$REQUIRED_VERSION" ]]; then
#       FOUND_PY=$(command -v $CAND)
#       break
#     fi
#   fi
# done

# # If we cannot find any working candidates, the user must install Python 3.12 manually
# if [[ -z "$FOUND_PY" ]]; then
#   echo "[ERROR] Python $REQUIRED_VERSION not found."
#   echo "Please install it and re-run this script."
#   exit 1
# fi

# # If we found a working candidate, we must use it
# echo "[INFO] Using Python interpreter: $FOUND_PY"
# echo "[INFO] Removing any existing virtual environment: $VENV_NAME"
# rm -rf "$VENV_NAME"
# echo "[INFO] Creating virtual environment: $VENV_NAME"
# $FOUND_PY -m venv "$VENV_NAME"

# # Defining all venv tools explicitly to prevent errors
# VENV_PY="bash $(pwd)/$VENV_NAME/bin/python"
# VENV_PIP="bash $(pwd)/$VENV_NAME/bin/pip"
# echo "[INFO] Activating environment..."
# source "$VENV_NAME/bin/activate"
# echo "[INFO] Python version in venv:"
# $VENV_PY --version
# # Ensuring the correct pip version is installed
# echo "[INFO] Ensuring pip, setuptools, and wheel are correct and isolated..."
# $VENV_PY -m ensurepip --upgrade
# $VENV_PIP install --upgrade pip setuptools wheel
# echo "[INFO] Forcing source build of pycryptodome for Python $REQUIRED_VERSION ABI..."
# $VENV_PIP uninstall -y pycryptodome || true
# PIP_NO_BINARY=:all: $VENV_PIP install --force-reinstall pycryptodome
# # Rebuilding pycryptodome from source to match Python 3.12's ABI to prevent errors
# echo "[INFO] Installing current project with [dev] extras..."
# $VENV_PIP install -e .[dev]
# echo "[INFO] Checking for required Python packages..."
# REQUIRED_PKGS=("paramiko" "pycryptodome" "pytest")
# for pkg in "${REQUIRED_PKGS[@]}"; do
#   if ! $VENV_PY -c "import $pkg" &>/dev/null; then
#     echo "[INSTALL] Installing missing package: $pkg"
#     $VENV_PIP install "$pkg"
#   fi
# done
# # Making sure that we have everything we need to pass all tests
# echo "[INFO] Checking for 'sftp' import..."
# if ! $VENV_PY -c "import sftp" &>/dev/null; then
#   echo "[WARN] Could not import 'sftp'. If it's a local file, check your PYTHONPATH."
#   echo "       If it's a package, please install it."
# fi
# echo "[INFO] SSH test info missing – attempting setup"
# if $VENV_PY -m tests.get_ssh_info 2>/dev/null; then
#   echo "[INFO] SSH info successfully configured."
# else
#   echo "[WARN] Could not configure SSH info – some tests may skip"
# fi

# # Printing a success message for the user to know that everything is ready to test
# echo "[SUCCESS] Virtual environment '$VENV_NAME' is ready."
# echo "[INFO] To use it, run: source $VENV_NAME/bin/activate"



# Setting various variables to their default states
set -euo pipefail
VENV_NAME=${1:-venv}
REQUIRED_VERSION="3.12"
FOUND_PY=""

# Trying to find common python3.12 binary candidate names and paths
echo "[INFO] Looking for Python $REQUIRED_VERSION..."
CANDIDATES=(
  python3.12
  /usr/bin/python3.12
  /usr/local/bin/python3.12
  "$HOME/.pyenv/versions/3.12.*/bin/python"
  "$HOME/.local/bin/python3.12"
)

# Iterating over the candidates to find a working one
for CAND in "${CANDIDATES[@]}"; do
  if command -v $CAND &>/dev/null; then
    VERSION=$($CAND -c 'import sys; print(".".join(map(str, sys.version_info[:2])))')
    if [[ "$VERSION" == "$REQUIRED_VERSION" ]]; then
      FOUND_PY=$(command -v $CAND)
      break
    fi
  fi
done

# If we cannot find any working candidates, the user must install Python 3.12 manually
if [[ -z "$FOUND_PY" ]]; then
  echo "[ERROR] Python $REQUIRED_VERSION not found."
  echo "Please install it and re-run this script."
  exit 1
fi

# If we found a working candidate, we must use it
echo "[INFO] Using Python interpreter: $FOUND_PY"
# Activating the correct venv
echo "[INFO] Removing any existing virtual environment: $VENV_NAME"
rm -rf "$VENV_NAME"
echo "[INFO] Creating virtual environment: $VENV_NAME"
$FOUND_PY -m venv "$VENV_NAME"
echo "[INFO] Activating environment..."
source "$VENV_NAME/bin/activate"
echo "[INFO] Python version in venv:"
python --version
# Ensuring the correct pip version is installed
echo "[INFO] Ensuring pip, setuptools, and wheel are correct and isolated..."
python -m ensurepip --upgrade
pip install --upgrade pip setuptools wheel
# Rebuilding pycryptodome from source to match Python 3.12 ABI, as there were errors beforehand
echo "[INFO] Forcing source build of pycryptodome for Python $REQUIRED_VERSION ABI..."
pip uninstall -y pycryptodome || true
PIP_NO_BINARY=:all: pip install --force-reinstall pycryptodome

# Installing the current package with all extra dependencies installed
echo "[INFO] Installing current project with [dev] extras..."
pip install -e .[dev]

# Printing a success message for the user to know that everything is ready to test
echo "[SUCCESS] Virtual environment '$VENV_NAME' is ready with:"
echo "  - Python $REQUIRED_VERSION"
echo "  - pycryptodome built from source"
echo "  - Local pip + dev dependencies"
echo "[INFO] After executing the source venv/bin/activate command, you can run the tox -r or pytest commands to test if this successful build actually works."
