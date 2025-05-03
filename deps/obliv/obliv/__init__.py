__all__ = ["hirb", "voram", "skipstash", "fstore", "mt_ssh_store", "ssh_info", "idstr"]

import sys

# Enforce supported Python version range at runtime (relaxed the limit in __init__.py and in pyproject.toml to allow for testing in the tox enviroment to work)
if not (3, 6) <= sys.version_info < (3, 14):
    raise RuntimeError(
        f"Tests for 'obliv' require Python >= 3.6 and < 3.14. "
        f"Your version is: {sys.version.split()[0]}"
    )
