import datetime

from sqlalchemy import Boolean, ForeignKey, Integer, String, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.infrastructure.database import Base


class Frame(Base):
    __tablename__ = "frames"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    project_id: Mapped[int] = mapped_column(ForeignKey("projects.id"), nullable=False)
    frame_index: Mapped[int] = mapped_column(Integer, nullable=False)
    is_keyframe: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    source_image_url: Mapped[str | None] = mapped_column(Text, nullable=True)
    created_at: Mapped[datetime.datetime] = mapped_column(default=datetime.datetime.utcnow)
    updated_at: Mapped[datetime.datetime] = mapped_column(
        default=datetime.datetime.utcnow, onupdate=datetime.datetime.utcnow
    )

    project: Mapped["Project"] = relationship(back_populates="frames")  # noqa: F821
    jobs: Mapped[list["Job"]] = relationship(back_populates="frame")  # noqa: F821
