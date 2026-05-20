"""
AES-GCM encryption for sensitive fields like source replication passwords.

Format: [12-byte nonce][ciphertext][16-byte auth tag]
Storage: combined as a single bytes blob in MySQL VARBINARY column.
"""

import base64
import os
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from dotenv import load_dotenv

load_dotenv()

NONCE_SIZE = 12  # 96 bits — AES-GCM standard

_encryption_key: bytes | None = None


def _get_key() -> bytes:
    """Load the AES key once and cache it. Raises if not set in production."""
    global _encryption_key
    if _encryption_key is not None:
        return _encryption_key

    key_b64 = os.getenv("ARGUS_ENCRYPTION_KEY")
    if not key_b64:
        # Dev fallback: generate ephemeral key. Data encrypted with this
        # cannot be decrypted after restart — fine for dev, broken for prod.
        print("WARNING: ARGUS_ENCRYPTION_KEY not set — using ephemeral dev key.")
        _encryption_key = AESGCM.generate_key(bit_length=256)
        return _encryption_key

    key = base64.b64decode(key_b64)
    if len(key) != 32:
        raise ValueError(
            f"ARGUS_ENCRYPTION_KEY must decode to exactly 32 bytes (got {len(key)})"
        )
    _encryption_key = key
    return _encryption_key


def encrypt(plaintext: str) -> bytes:
    """Encrypt a string. Returns: nonce || ciphertext || tag (all in one blob)."""
    aesgcm = AESGCM(_get_key())
    nonce = os.urandom(NONCE_SIZE)
    ct = aesgcm.encrypt(nonce, plaintext.encode("utf-8"), associated_data=None)
    return nonce + ct


def decrypt(blob: bytes) -> str:
    """Decrypt a blob produced by encrypt(). Returns the original string."""
    if len(blob) < NONCE_SIZE:
        raise ValueError("Encrypted blob too short")
    nonce = blob[:NONCE_SIZE]
    ct = blob[NONCE_SIZE:]
    aesgcm = AESGCM(_get_key())
    pt = aesgcm.decrypt(nonce, ct, associated_data=None)
    return pt.decode("utf-8")