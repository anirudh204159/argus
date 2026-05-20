"""Rate limiter for auth endpoints.

Uses in-memory counters by default — fine for single-instance deployments.
For multi-instance, swap storage_uri to Redis: "redis://localhost:6379".

Set ARGUS_DISABLE_RATE_LIMIT=1 to disable in tests.
"""

import os

from slowapi import Limiter
from slowapi.util import get_remote_address

_disabled = os.getenv("ARGUS_DISABLE_RATE_LIMIT") == "1"

limiter = Limiter(
    key_func=get_remote_address,
    enabled=not _disabled,
)