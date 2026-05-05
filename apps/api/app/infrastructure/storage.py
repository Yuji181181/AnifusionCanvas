import io

import boto3
from botocore.config import Config as BotoConfig

from app.config import settings


class R2Storage:
    def __init__(self) -> None:
        self._client = boto3.client(
            "s3",
            endpoint_url=settings.r2_endpoint_url,
            aws_access_key_id=settings.r2_access_key_id,
            aws_secret_access_key=settings.r2_secret_access_key,
            region_name=settings.r2_region,
            config=BotoConfig(signature_version="s3v4"),
        )
        self._bucket = settings.r2_bucket
        self._public_base_url = settings.r2_public_base_url.rstrip("/")

    def _object_url(self, key: str) -> str:
        if self._public_base_url:
            return f"{self._public_base_url}/{key}"
        return key

    def upload_bytes(self, key: str, data: bytes, content_type: str = "image/png") -> str:
        self._client.upload_fileobj(
            io.BytesIO(data),
            self._bucket,
            key,
            ExtraArgs={"ContentType": content_type},
        )
        return self._object_url(key)

    def upload_file(self, key: str, file_path: str, content_type: str = "image/png") -> str:
        self._client.upload_file(
            file_path,
            self._bucket,
            key,
            ExtraArgs={"ContentType": content_type},
        )
        return self._object_url(key)

    def download_bytes(self, key: str) -> bytes:
        resp = self._client.get_object(Bucket=self._bucket, Key=key)
        return resp["Body"].read()

    def delete(self, key: str) -> None:
        self._client.delete_object(Bucket=self._bucket, Key=key)

    def generate_presigned_url(self, key: str, expires_in: int = 3600) -> str:
        return self._client.generate_presigned_url(
            "get_object",
            Params={"Bucket": self._bucket, "Key": key},
            ExpiresIn=expires_in,
        )


r2_storage = R2Storage()
