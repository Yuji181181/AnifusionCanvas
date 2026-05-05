import os
import tempfile

import httpx
import replicate

from app.config import settings

TOONCRAFTER_MODEL = "fofr/tooncrafter"
SD_INPAINT_MODEL = "lucataco/sdxl-inpainting"


async def run_tooncrafter(
    image_1_url: str,
    image_2_url: str,
    prompt: str = "",
    negative_prompt: str = "",
    max_width: int = 512,
    max_height: int = 512,
    seed: int | None = None,
) -> list[str]:
    """Run ToonCrafter inference via Replicate API.
    Returns a list of frame image URLs (extracted from MP4 output).
    """
    input_params: dict = {
        "image_1": image_1_url,
        "image_2": image_2_url,
        "prompt": prompt,
        "negative_prompt": negative_prompt,
        "max_width": max_width,
        "max_height": max_height,
    }
    if seed is not None:
        input_params["seed"] = seed

    output = await replicate.async_run(TOONCRAFTER_MODEL, input=input_params)

    # output is a list of URI strings; first element is the MP4 URL
    video_url = output[0] if isinstance(output, list) else output
    return await _extract_frames_from_video(video_url)


async def run_inpainting(
    image_url: str,
    mask_url: str,
    prompt: str,
    negative_prompt: str = "",
    strength: float = 0.7,
    steps: int = 20,
    guidance_scale: float = 8.0,
    seed: int | None = None,
) -> str:
    """Run SDXL Inpainting via Replicate API.
    Returns the URL of the inpainted image.
    """
    input_params: dict = {
        "image": image_url,
        "mask": mask_url,
        "prompt": prompt,
        "negative_prompt": negative_prompt,
        "strength": strength,
        "steps": steps,
        "guidance_scale": guidance_scale,
    }
    if seed is not None:
        input_params["seed"] = seed

    output = await replicate.async_run(SD_INPAINT_MODEL, input=input_params)

    # output is a list of image URLs
    return output[0] if isinstance(output, list) else output


async def _extract_frames_from_video(video_url: str) -> list[str]:
    """Download MP4 from URL, extract frames as PNGs using ffmpeg, upload to R2."""
    import asyncio

    from app.infrastructure.storage import r2_storage

    with tempfile.TemporaryDirectory() as tmpdir:
        video_path = f"{tmpdir}/input.mp4"

        # Download video
        async with httpx.AsyncClient(timeout=120) as client:
            resp = await client.get(video_url)
            resp.raise_for_status()
            with open(video_path, "wb") as f:
                f.write(resp.content)

        # Extract frames with ffmpeg
        frames_pattern = f"{tmpdir}/frame_%04d.png"
        proc = await asyncio.create_subprocess_exec(
            "ffmpeg", "-i", video_path, "-q:v", "2", frames_pattern,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )
        await proc.wait()

        # Upload each frame to R2
        import glob

        frame_files = sorted(glob.glob(f"{tmpdir}/frame_*.png"))
        frame_urls: list[str] = []
        for i, frame_path in enumerate(frame_files):
            key = f"generated/{os.urandom(8).hex()}/frame_{i:04d}.png"
            url = r2_storage.upload_file(key, frame_path, content_type="image/png")
            frame_urls.append(url)

    return frame_urls
