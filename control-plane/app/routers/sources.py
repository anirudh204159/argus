from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Source, User
from app.schemas import SourceCreate, SourceUpdate, SourceOut
from app.security import get_current_user


router = APIRouter(prefix="/sources", tags=["sources"])


@router.post("", response_model=SourceOut, status_code=status.HTTP_201_CREATED)
def create_source(
    payload: SourceCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Register a new MySQL source for the current user."""
    source = Source(
        user_id=current_user.id,
        name=payload.name,
        host=payload.host,
        port=payload.port,
        database_name=payload.database_name,
        replication_user=payload.replication_user,
        replication_password_enc=payload.replication_password.encode("utf-8"),
    )
    db.add(source)
    try:
        db.commit()
    except IntegrityError:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"You already have a source named '{payload.name}'",
        )
    db.refresh(source)
    return source


@router.get("", response_model=list[SourceOut])
def list_sources(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """List all sources owned by the current user."""
    return db.query(Source).filter(Source.user_id == current_user.id).all()


@router.get("/{source_id}", response_model=SourceOut)
def get_source(
    source_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Get a single source by ID."""
    source = (
        db.query(Source)
        .filter(Source.id == source_id, Source.user_id == current_user.id)
        .first()
    )
    if source is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Source not found")
    return source


@router.patch("/{source_id}", response_model=SourceOut)
def update_source(
    source_id: int,
    payload: SourceUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Update fields on a source. Only fields included in the body are changed."""
    source = (
        db.query(Source)
        .filter(Source.id == source_id, Source.user_id == current_user.id)
        .first()
    )
    if source is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Source not found")

    updates = payload.model_dump(exclude_unset=True)

    if "replication_password" in updates:
        source.replication_password_enc = updates.pop("replication_password").encode("utf-8")

    for field, value in updates.items():
        setattr(source, field, value)

    try:
        db.commit()
    except IntegrityError:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="A source with that name already exists",
        )
    db.refresh(source)
    return source


@router.delete("/{source_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_source(
    source_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Delete a source. All related subscriptions and the checkpoint are cascaded."""
    source = (
        db.query(Source)
        .filter(Source.id == source_id, Source.user_id == current_user.id)
        .first()
    )
    if source is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Source not found")

    db.delete(source)
    db.commit()
    return None