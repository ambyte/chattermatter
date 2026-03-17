#!/usr/bin/env python3
"""
Mattermost License Generator

Generates RSA key pair and creates a signed Mattermost license.
Compatible with Mattermost license validation (RSA 2048, SHA512, PKCS1v15).

Usage:
    python generate_license.py                    # Generate new keys + license
    python generate_license.py --keys-only         # Generate keys only
    python generate_license.py --license-only     # Sign license with existing key
"""

import argparse
import base64
import hashlib
import json
import secrets
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path

# RSA 2048 for Mattermost compatibility (256-byte signature)
RSA_KEY_SIZE = 2048

# Mattermost IDs must be exactly 26 alphanumeric chars (IsValidId)
MATTERMOST_ID_ALPHABET = "abcdefghijklmnopqrstuvwxyz0123456789"


def mattermost_id() -> str:
    """Generate Mattermost-compatible ID (26 chars, alphanumeric)."""
    return "".join(secrets.choice(MATTERMOST_ID_ALPHABET) for _ in range(26))


def generate_rsa_keypair() -> tuple:
    """Generate RSA key pair. Returns (private_key, public_key) as PEM bytes."""
    try:
        from cryptography.hazmat.primitives.asymmetric import rsa
        from cryptography.hazmat.primitives import serialization
        from cryptography.hazmat.backends import default_backend
    except ImportError:
        print("Error: Install cryptography: pip install cryptography", file=sys.stderr)
        sys.exit(1)

    private_key = rsa.generate_private_key(
        public_exponent=65537,
        key_size=RSA_KEY_SIZE,
        backend=default_backend(),
    )

    private_pem = private_key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    )

    public_key = private_key.public_key()
    public_pem = public_key.public_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PublicFormat.SubjectPublicKeyInfo,
    )

    return private_pem, public_pem


def sign_license_with_cryptography(plaintext: bytes, private_pem: bytes) -> bytes:
    """Sign license using cryptography library."""
    from cryptography.hazmat.primitives import serialization, hashes
    from cryptography.hazmat.primitives.asymmetric import padding
    from cryptography.hazmat.backends import default_backend

    private_key = serialization.load_pem_private_key(
        private_pem, password=None, backend=default_backend()
    )

    signature = private_key.sign(
        plaintext,
        padding.PKCS1v15(),
        hashes.SHA512(),
    )
    return signature


def sign_license_with_stdlib(plaintext: bytes, private_pem: bytes) -> bytes:
    """Sign license using standard library only (no external deps)."""
    import ssl

    # Use ssl to load PEM - we need the raw key for signing
    # Actually, Python's ssl/crypto in stdlib is limited. Let's use cryptography.
    raise NotImplementedError("Use 'pip install cryptography' for license signing")


def create_license_json(
    users: int = 20000,
    company: str = "My Company",
    email: str = "admin@example.com",
    name: str = "Admin",
    sku: str = "Professional",
    sku_short: str = "professional",
    is_trial: bool = False,
    days_valid: int = 365,
) -> dict:
    """Create license JSON structure compatible with Mattermost."""
    now = datetime.now(timezone.utc)
    issued_at = int(now.timestamp() * 1000)
    starts_at = issued_at
    expires_at = int((now + timedelta(days=days_valid)).timestamp() * 1000)

    license_id = mattermost_id()
    customer_id = mattermost_id()

    return {
        "id": license_id,
        "issued_at": issued_at,
        "starts_at": starts_at,
        "expires_at": expires_at,
        "sku_name": sku,
        "sku_short_name": sku_short,
        "customer": {
            "id": customer_id,
            "name": name,
            "email": email,
            "company": company,
        },
        "features": {
            "users": users,
            "ldap": True,
            "ldap_groups": False,
            "mfa": True,
            "google_oauth": False,
            "office365_oauth": False,
            "openid": True,
            "compliance": False,
            "cluster": True,
            "metrics": True,
            "mhpns": True,
            "saml": True,
            "elastic_search": False,
            "announcement": True,
            "theme_management": True,
            "email_notification_contents": True,
            "data_retention": True,
            "message_export": True,
            "custom_permissions_schemes": True,
            "custom_terms_of_service": True,
            "guest_accounts": True,
            "guest_accounts_permissions": True,
            "id_loaded": True,
            "lock_teammate_name_display": True,
            "enterprise_plugins": True,
            "advanced_logging": True,
            "cloud": False,
            "shared_channels": True,
            "remote_cluster_service": False,
            "outgoing_oauth_connections": True,
            "future_features": True,
        },
        "is_trial": is_trial,
        "is_gov_sku": False,
    }


def sign_license(plaintext: bytes, private_pem: bytes) -> bytes:
    """Sign license plaintext with private key. Returns 256-byte signature."""
    return sign_license_with_cryptography(plaintext, private_pem)


def create_signed_license(license_data: dict, private_pem: bytes) -> str:
    """
    Create signed license in Mattermost format.
    Format: base64(plaintext_json + signature_256bytes)
    """
    plaintext = json.dumps(license_data, separators=(",", ":")).encode("utf-8")
    signature = sign_license(plaintext, private_pem)

    if len(signature) != 256:
        raise ValueError(f"RSA 2048 signature must be 256 bytes, got {len(signature)}")

    signed = plaintext + signature
    return base64.b64encode(signed).decode("ascii")


def main():
    parser = argparse.ArgumentParser(description="Mattermost License Generator")
    parser.add_argument(
        "--output-dir",
        default=".",
        help="Directory for output files (default: current)",
    )
    parser.add_argument(
        "--keys-only",
        action="store_true",
        help="Only generate RSA key pair, do not create license",
    )
    parser.add_argument(
        "--license-only",
        action="store_true",
        help="Only create license using existing private key",
    )
    parser.add_argument(
        "--private-key",
        default="license-private.pem",
        help="Path to private key file (default: license-private.pem)",
    )
    parser.add_argument(
        "--users",
        type=int,
        default=20000,
        help="Number of licensed users (default: 20000)",
    )
    parser.add_argument(
        "--company",
        default="My Company",
        help="Company name (default: My Company)",
    )
    parser.add_argument(
        "--email",
        default="admin@example.com",
        help="Admin email (default: admin@example.com)",
    )
    parser.add_argument(
        "--days",
        type=int,
        default=365,
        help="License validity in days (default: 365)",
    )
    parser.add_argument(
        "--sku",
        default="Professional",
        choices=["Professional", "Enterprise", "E10", "E20"],
        help="SKU name (default: Professional)",
    )
    args = parser.parse_args()

    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    private_key_path = output_dir / args.private_key
    public_key_path = output_dir / "license-public-key.txt"
    license_path = output_dir / "mattermost.mattermost-license"

    if args.license_only:
        if not private_key_path.exists():
            print(f"Error: Private key not found: {private_key_path}", file=sys.stderr)
            sys.exit(1)
        private_pem = private_key_path.read_bytes()
    else:
        # Generate new key pair
        print("Generating RSA key pair...")
        private_pem, public_pem = generate_rsa_keypair()

        private_key_path.write_bytes(private_pem)
        public_key_path.write_bytes(public_pem)
        print(f"  Private key: {private_key_path}")
        print(f"  Public key:  {public_key_path}")

        if args.keys_only:
            print("\nDone. Use --license-only to create a license later.")
            return

    # Create and sign license
    sku_short_map = {
        "Professional": "professional",
        "Enterprise": "enterprise",
        "E10": "E10",
        "E20": "E20",
    }
    license_data = create_license_json(
        users=args.users,
        company=args.company,
        email=args.email,
        days_valid=args.days,
        sku=args.sku,
        sku_short=sku_short_map[args.sku],
    )

    print("\nCreating signed license...")
    signed_license = create_signed_license(license_data, private_pem)

    license_path.write_text(signed_license, encoding="utf-8")
    print(f"  License file: {license_path}")

    print("\n" + "=" * 60)
    print("To use with Mattermost (test environment):")
    print("  1. Copy license-public-key.txt to server/channels/utils/")
    print("     as license-public-key-test.txt")
    print("  2. Set MM_SERVICEENVIRONMENT=test")
    print("  3. Place mattermost.mattermost-license in config/ or upload via UI")
    print("=" * 60)


if __name__ == "__main__":
    main()
