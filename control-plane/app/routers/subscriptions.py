import secrets

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Source, Subscription, User
from app.schemas import (
    SubscriptionCreate,
    SubscriptionUpdate,
    SubscriptionOut,
    SubscriptionCreated,
)
from app.security import get_current_user


router = APIRouter(prefix="/subscriptions", tags=["subscriptions"])


ALLOWED_OPERATIONS = {"INSERT", "UPDATE", "DELETE"}


def _validate_operations(operations: list[str]) -> None:
    invalid = set(operations) - ALLOWED_OPERATIONS
    if invalid:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Invalid operations: {sorted(invalid)}. Allowed: {sorted(ALLOWED_OPERATIONS)}",
        )


@router.post("", response_model=SubscriptionCreated, status_code=status.HTTP_201_CREATED)
def create_subscription(
    payload: SubscriptionCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Create a new subscription. The HMAC secret is returned only once."""
    _validate_operations(payload.operations)

    # Verify the source belongs to the current user
    source = (
        db.query(Source)
        .filter(Source.id == payload.source_id, Source.user_id == current_user.id)
        .first()
    )
    if source is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Source not found",
        )

    subscription = Subscription(
        user_id=current_user.id,
        source_id=payload.source_id,
        name=payload.name,
        tables=payload.tables,
        operations=payload.operations,
        webhook_url=payload.webhook_url,
        hmac_secret=secrets.token_urlsafe(32),
        retry_max=payload.retry_max,
    )
    db.add(subscription)
    try:
        db.commit()
    except IntegrityError:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"You already have a subscription named '{payload.name}'",
        )
    db.refresh(subscription)
    return subscription


@router.get("", response_model=list[SubscriptionOut])
def list_subscriptions(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """List all subscriptions owned by the current user."""
    return (
        db.query(Subscription)
        .filter(Subscription.user_id == current_user.id)
        .all()
    )


@router.get("/{subscription_id}", response_model=SubscriptionOut)
def get_subscription(
    subscription_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Get a single subscription by ID."""
    subscription = (
        db.query(Subscription)
        .filter(
            Subscription.id == subscription_id,
            Subscription.user_id == current_user.id,
        )
        .first()
    )
    if subscription is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Subscription not found",
        )
    return subscription


@router.patch("/{subscription_id}", response_model=SubscriptionOut)
def update_subscription(
    subscription_id: int,
    payload: SubscriptionUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Update fields on a subscription. Only fields included in the body are changed."""
    subscription = (
        db.query(Subscription)
        .filter(
            Subscription.id == subscription_id,
            Subscription.user_id == current_user.id,
        )
        .first()
    )
    if subscription is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Subscription not found",
        )

    updates = payload.model_dump(exclude_unset=True)

    if "operations" in updates:
        _validate_operations(updates["operations"])

    for field, value in updates.items():
        setattr(subscription, field, value)

    try:
        db.commit()
    except IntegrityError:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="A subscription with that name already exists",
        )
    db.refresh(subscription)
    return subscription


@router.delete("/{subscription_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_subscription(
    subscription_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """Delete a subscription. Related delivery logs and DLQ entries are cascaded."""
    subscription = (
        db.query(Subscription)
        .filter(
            Subscription.id == subscription_id,
            Subscription.user_id == current_user.id,
        )
        .first()
    )
    if subscription is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Subscription not found",
        )

    db.delete(subscription)
    db.commit()
    return None