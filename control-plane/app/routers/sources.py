from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Source, User
from app.redis_client import publish_config_change
from app.schemas import SourceCreate, SourceOut, SourceUpdate
from app.security import get_current_user

router = APIRouter(prefix="/sources", tags=["sources"])


@router.post("", response_model=SourceOut, status_code=status.HTTP_201_CREATED)
def create_source(
    payload: SourceCreate,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    source = Source(
        user_id=user.id,
        name=payload.name,
        host=payload.host,
        port=payload.port,
        database_name=payload.database_name,
        replication_user=payload.replication_user,
        replication_password_enc=payload.replication_password.encode(),
        status="disconnected",
    )
    db.add(source)
    try:
        db.commit()
    except IntegrityError:
        db.rollback()
        raise HTTPException(status_code=409, detail="Source with that name already exists")
    db.refresh(source)
    return source


@router.get("", response_model=list[SourceOut])
def list_sources(
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    return db.execute(
        select(Source).where(Source.user_id == user.id)
    ).scalars().all()


@router.get("/{source_id}", response_model=SourceOut)
def get_source(
    source_id: int,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    source = db.execute(
        select(Source).where(Source.id == source_id, Source.user_id == user.id)
    ).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="Source not found")
    return source


@router.patch("/{source_id}", response_model=SourceOut)
def update_source(
    source_id: int,
    payload: SourceUpdate,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    source = db.execute(
        select(Source).where(Source.id == source_id, Source.user_id == user.id)
    ).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="Source not found")

    update_data = payload.model_dump(exclude_unset=True)
    if "replication_password" in update_data:
        update_data["replication_password_enc"] = update_data.pop("replication_password").encode()
    for field, value in update_data.items():
        setattr(source, field, value)

    db.commit()
    db.refresh(source)

    publish_config_change("source_updated", source_id)
    return source


@router.delete("/{source_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_source(
    source_id: int,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    source = db.execute(
        select(Source).where(Source.id == source_id, Source.user_id == user.id)
    ).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="Source not found")

    db.delete(source)
    db.commit()

    publish_config_change("source_deleted", source_id)