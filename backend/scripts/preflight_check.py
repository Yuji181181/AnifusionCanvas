from __future__ import annotations

import os
import shutil
import subprocess
from dataclasses import dataclass

import boto3
from dotenv import load_dotenv
from huggingface_hub import HfApi
from sqlalchemy import create_engine, text


@dataclass
class CheckResult:
    name: str
    ok: bool
    message: str


def check_command(name: str) -> CheckResult:
    path = shutil.which(name)
    if path:
        return CheckResult(name=f"command:{name}", ok=True, message=path)
    return CheckResult(name=f"command:{name}", ok=False, message="not found")


def check_required_env() -> list[CheckResult]:
    required = [
        "DATABASE_URL",
        "R2_ACCOUNT_ID",
        "R2_ACCESS_KEY_ID",
        "R2_SECRET_ACCESS_KEY",
        "R2_BUCKET",
        "R2_ENDPOINT_URL",
    ]

    results: list[CheckResult] = []
    for key in required:
        value = os.getenv(key)
        results.append(
            CheckResult(
                name=f"env:{key}", ok=bool(value), message="set" if value else "missing"
            )
        )
    return results


def check_tidb() -> CheckResult:
    database_url = os.getenv("DATABASE_URL", "")
    if not database_url:
        return CheckResult(name="tidb", ok=False, message="DATABASE_URL is missing")

    if database_url.startswith("mysql://"):
        return CheckResult(
            name="tidb",
            ok=False,
            message=(
                "Use mysql+pymysql://... (not mysql://...). "
                "Current URL points to MySQLdb driver, which is not installed."
            ),
        )

    try:
        engine = create_engine(database_url, pool_pre_ping=True)
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
        return CheckResult(name="tidb", ok=True, message="connected")
    except Exception as exc:  # noqa: BLE001
        return CheckResult(name="tidb", ok=False, message=str(exc))


def check_r2() -> CheckResult:
    endpoint_url = os.getenv("R2_ENDPOINT_URL")
    access_key = os.getenv("R2_ACCESS_KEY_ID")
    secret_key = os.getenv("R2_SECRET_ACCESS_KEY")
    bucket = os.getenv("R2_BUCKET")

    if not all([endpoint_url, access_key, secret_key, bucket]):
        return CheckResult(name="r2", ok=False, message="R2_* env vars are incomplete")

    try:
        client = boto3.client(
            "s3",
            endpoint_url=endpoint_url,
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            region_name=os.getenv("R2_REGION", "auto"),
        )
        client.head_bucket(Bucket=bucket)
        return CheckResult(name="r2", ok=True, message="bucket reachable")
    except Exception as exc:  # noqa: BLE001
        return CheckResult(name="r2", ok=False, message=str(exc))


def check_hf() -> CheckResult:
    token = os.getenv("HF_TOKEN")
    if not token:
        return CheckResult(name="huggingface", ok=False, message="HF_TOKEN is missing")

    try:
        profile = HfApi().whoami(token=token)
        user_name = profile.get("name", "unknown")
        return CheckResult(
            name="huggingface", ok=True, message=f"authenticated as {user_name}"
        )
    except Exception as exc:  # noqa: BLE001
        return CheckResult(name="huggingface", ok=False, message=str(exc))


def check_modal_auth() -> CheckResult:
    if not shutil.which("modal"):
        return CheckResult(name="modal", ok=False, message="modal CLI not found")

    try:
        proc = subprocess.run(
            ["modal", "profile", "list"],
            check=False,
            capture_output=True,
            text=True,
        )
        if proc.returncode == 0:
            return CheckResult(
                name="modal", ok=True, message="profile command succeeded"
            )
        return CheckResult(
            name="modal", ok=False, message=proc.stderr.strip() or proc.stdout.strip()
        )
    except Exception as exc:  # noqa: BLE001
        return CheckResult(name="modal", ok=False, message=str(exc))


def print_results(results: list[CheckResult]) -> int:
    failed = 0
    for item in results:
        status = "OK" if item.ok else "NG"
        print(f"[{status}] {item.name}: {item.message}")
        if not item.ok:
            failed += 1

    if failed:
        print(f"\nPreflight failed: {failed} checks failed")
        return 1

    print("\nPreflight passed: all checks are green")
    return 0


def main() -> int:
    load_dotenv()

    results: list[CheckResult] = []
    for cmd in ["uv", "bun", "modal", "cloudflared", "aws", "ffmpeg"]:
        results.append(check_command(cmd))

    results.extend(check_required_env())
    results.append(check_modal_auth())
    results.append(check_hf())
    results.append(check_tidb())
    results.append(check_r2())

    return print_results(results)


if __name__ == "__main__":
    raise SystemExit(main())
