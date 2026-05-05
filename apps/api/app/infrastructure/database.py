import ssl
from collections.abc import AsyncGenerator
from urllib.parse import urlparse, urlunparse

from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.orm import DeclarativeBase

from app.config import settings


def _build_async_db_url() -> str:
    url = settings.database_url.replace("mysql+pymysql", "mysql+aiomysql", 1)
    parsed = urlparse(url)
    query = parsed.query.replace("ssl_verify_cert=true", "").rstrip("&").lstrip("&")
    clean = parsed._replace(query=query)
    return urlunparse(clean)


ssl_ctx = ssl.create_default_context()
db_url = _build_async_db_url()
engine = create_async_engine(
    db_url,
    pool_pre_ping=True,
    echo=False,
    connect_args={"ssl": ssl_ctx},
)
async_session = async_sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)


class Base(DeclarativeBase):
    pass


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session() as session:
        yield session
