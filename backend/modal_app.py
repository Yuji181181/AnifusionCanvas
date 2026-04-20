from __future__ import annotations

import modal

APP_NAME = "anifusion-canvas"

image = (
    modal.Image.debian_slim(python_version="3.11")
    .apt_install("ffmpeg")
    .pip_install(
        "torch",
        "torchvision",
        "diffusers",
        "transformers",
        "accelerate",
        "safetensors",
        "opencv-python-headless",
    )
)

app = modal.App(APP_NAME)


@app.function(image=image, gpu="A10G", timeout=1800)
def healthcheck() -> dict[str, str]:
    return {"status": "ok", "app": APP_NAME}
