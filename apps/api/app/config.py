from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    model_config = {"env_file": ".env", "env_file_encoding": "utf-8", "extra": "ignore"}

    # Database
    database_url: str = ""

    # Cloudflare R2
    r2_account_id: str = ""
    r2_access_key_id: str = ""
    r2_secret_access_key: str = ""
    r2_bucket: str = ""
    r2_public_base_url: str = ""
    r2_endpoint_url: str = ""
    r2_region: str = "auto"

    # Hugging Face
    hf_token: str = ""
    model_cache_dir: str = "./artifacts/models"
    tooncrafter_model_id: str = "Doubiiu/ToonCrafter"
    sd_inpaint_model_id: str = "diffusers/stable-diffusion-xl-1.0-inpainting-0.1"

    # Modal
    modal_profile: str = ""
    modal_app_name: str = ""
    modal_endpoint: str = ""


settings = Settings()
