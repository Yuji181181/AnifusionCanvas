from __future__ import annotations

import os
import shutil
import subprocess
from pathlib import Path

from dotenv import load_dotenv
from huggingface_hub import HfApi, hf_hub_url


def _env_int(name: str, default: int) -> int:
    raw = os.getenv(name, str(default)).strip()
    try:
        value = int(raw)
        if value < 1:
            return default
        return value
    except ValueError:
        return default


def _env_str(name: str) -> str | None:
    value = os.getenv(name, "").strip()
    return value or None


def _download_one_file_with_aria2(
    *,
    url: str,
    target_file: Path,
    token: str | None,
    split: int,
    connect_per_server: int,
    max_concurrent_downloads: int,
    retry_wait_seconds: int,
    max_tries: int,
    timeout_seconds: int,
    max_download_limit: str | None,
) -> None:
    target_file.parent.mkdir(parents=True, exist_ok=True)

    command = [
        "aria2c",
        "--continue=true",
        "--allow-overwrite=true",
        "--auto-file-renaming=false",
        "--file-allocation=none",
        "--min-split-size=1M",
        f"--max-connection-per-server={connect_per_server}",
        f"--split={split}",
        f"--max-concurrent-downloads={max_concurrent_downloads}",
        f"--retry-wait={retry_wait_seconds}",
        f"--max-tries={max_tries}",
        f"--connect-timeout={timeout_seconds}",
        f"--timeout={timeout_seconds}",
        "--disable-ipv6=true",
        "--dir",
        str(target_file.parent),
        "--out",
        target_file.name,
        url,
    ]

    if max_download_limit:
        command.extend(["--max-overall-download-limit", max_download_limit])

    if token:
        command.extend(["--header", f"Authorization: Bearer {token}"])

    proc = subprocess.run(command, check=False)
    if proc.returncode != 0:
        raise RuntimeError(f"aria2c failed for {target_file} (exit={proc.returncode})")


def download_if_configured(
    model_id: str,
    target_dir: Path,
    token: str | None,
    split: int,
    connect_per_server: int,
    max_concurrent_downloads: int,
    retry_wait_seconds: int,
    max_tries: int,
    timeout_seconds: int,
    max_download_limit: str | None,
) -> None:
    api = HfApi(token=token)
    revision = os.getenv("HF_MODEL_REVISION", "main")

    print(f"Listing files for {model_id}@{revision}")
    files = api.list_repo_files(repo_id=model_id, repo_type="model", revision=revision)
    if not files:
        raise RuntimeError(f"No files found in {model_id}@{revision}")

    print(f"Downloading {len(files)} files from {model_id} -> {target_dir}")
    for index, file_path in enumerate(files, start=1):
        output_path = target_dir / file_path
        url = hf_hub_url(
            repo_id=model_id,
            filename=file_path,
            repo_type="model",
            revision=revision,
        )
        print(f"[{index}/{len(files)}] {file_path}")
        _download_one_file_with_aria2(
            url=url,
            target_file=output_path,
            token=token,
            split=split,
            connect_per_server=connect_per_server,
            max_concurrent_downloads=max_concurrent_downloads,
            retry_wait_seconds=retry_wait_seconds,
            max_tries=max_tries,
            timeout_seconds=timeout_seconds,
            max_download_limit=max_download_limit,
        )


def _ensure_aria2_installed() -> None:
    if shutil.which("aria2c"):
        return
    raise RuntimeError(
        "aria2c is not installed. Install it first (e.g. apt install aria2)."
    )


def main() -> int:
    load_dotenv()

    _ensure_aria2_installed()

    # Wi-Fiが不安定な環境向けに、デフォルトを保守的にする。
    split = _env_int("ARIA2_SPLIT", 4)
    connect_per_server = _env_int("ARIA2_MAX_CONNECTION_PER_SERVER", 4)
    max_concurrent_downloads = _env_int("ARIA2_MAX_CONCURRENT_DOWNLOADS", 1)
    retry_wait_seconds = _env_int("ARIA2_RETRY_WAIT", 5)
    max_tries = _env_int("ARIA2_MAX_TRIES", 20)
    timeout_seconds = _env_int("ARIA2_TIMEOUT", 30)
    max_download_limit = _env_str("ARIA2_MAX_OVERALL_DOWNLOAD_LIMIT")

    cache_root = Path(os.getenv("MODEL_CACHE_DIR", "./artifacts/models")).resolve()
    cache_root.mkdir(parents=True, exist_ok=True)

    hf_token = os.getenv("HF_TOKEN")
    tooncrafter_model_id = os.getenv("TOONCRAFTER_MODEL_ID", "").strip()
    sd_inpaint_model_id = os.getenv("SD_INPAINT_MODEL_ID", "").strip()

    if not tooncrafter_model_id and not sd_inpaint_model_id:
        print(
            "No model IDs configured. Set TOONCRAFTER_MODEL_ID and/or SD_INPAINT_MODEL_ID."
        )
        return 1

    print(
        "aria2 settings: "
        f"split={split} "
        f"max_connection_per_server={connect_per_server} "
        f"max_concurrent_downloads={max_concurrent_downloads} "
        f"retry_wait={retry_wait_seconds} "
        f"max_tries={max_tries} "
        f"timeout={timeout_seconds} "
        f"max_overall_download_limit={max_download_limit or 'none'}"
    )

    if tooncrafter_model_id:
        download_if_configured(
            tooncrafter_model_id,
            cache_root / "tooncrafter",
            hf_token,
            split,
            connect_per_server,
            max_concurrent_downloads,
            retry_wait_seconds,
            max_tries,
            timeout_seconds,
            max_download_limit,
        )

    if sd_inpaint_model_id:
        download_if_configured(
            sd_inpaint_model_id,
            cache_root / "sd-inpaint",
            hf_token,
            split,
            connect_per_server,
            max_concurrent_downloads,
            retry_wait_seconds,
            max_tries,
            timeout_seconds,
            max_download_limit,
        )

    print("Model download completed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
