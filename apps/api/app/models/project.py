import datetime
from enum import StrEnum

from sqlalchemy import ForeignKey, Integer, String, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.infrastructure.database import Base


class ProjectStatus(StrEnum):
    DRAFT = "draft"
    ACTIVE = "active"
    COMPLETED = "completed"


class Project(Base):
    __tablename__ = "projects"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id"), nullable=False)
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(Text, nullable=True)
    width: Mapped[int] = mapped_column(Integer, nullable=False, default=512)
    height: Mapped[int] = mapped_column(Integer, nullable=False, default=512)
    fps: Mapped[int] = mapped_column(Integer, nullable=False, default=8)
    status: Mapped[str] = mapped_column(String(50), nullable=False, default=ProjectStatus.DRAFT)
    created_at: Mapped[datetime.datetime] = mapped_column(default=datetime.datetime.utcnow)
    updated_at: Mapped[datetime.datetime] = mapped_column(
        default=datetime.datetime.utcnow, onupdate=datetime.datetime.utcnow
    )

    user: Mapped["User"] = relationship(back_populates="projects")  # noqa: F821
    frames: Mapped[list["Frame"]] = relationship(back_populates="project", cascade="all, delete-orphan", order_by="Frame.frame_index")  # noqa: F821
    jobs: Mapped[list["Job"]] = relationship(back_populates="project", cascade="all, delete-orphan")  # noqa: F821
