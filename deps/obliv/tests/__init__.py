# tests/__init__.py Ismail Ahmed: It is the test suite package for the obliv library
import sys
"""It is the test suite package for the obliv library."""

# Enforce supported Python version range at runtime (relaxed the limit in __init__.py and in pyproject.toml to allow for testing in the tox enviroment to work)
if not (3, 6) <= sys.version_info < (3, 14):
    raise RuntimeError(
        f"obliv only supports Python >= 3.6 and < 3.14. "
        f"Your version is: {sys.version.split()[0]}"
    )