import secrets

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Source, Subscription, User
from app.redis_client import publish_config_change
from app.schemas import (
    SubscriptionCreate,
    SubscriptionCreated,
    SubscriptionOut,
    SubscriptionUpdate,
)
from app.security import get_current_user

router = APIRouter(prefix="/subscriptions", tags=["subscriptions"])


@router.post("", response_model=SubscriptionCreated, status_code=status.HTTP_201_CREATED)
def create_subscription(
    payload: SubscriptionCreate,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    # Verify the source belongs to this user
    source = db.execute(
        select(Source).where(Source.id == payload.source_id, Source.user_id == user.id)
    ).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="Source not found")

    sub = Subscription(
        user_id=user.id,
        source_id=payload.source_id,
        name=payload.name,
        tables=payload.tables,
        operations=payload.operations,
        webhook_url=str(payload.webhook_url),
        hmac_secret=secrets.token_urlsafe(32),
        retry_max=payload.retry_max,
        active=True,
    )
    db.add(sub)
    db.commit()
    db.refresh(sub)

    publish_config_change("subscription_created", sub.id)
    return sub


@router.get("", response_model=list[SubscriptionOut])
def list_subscriptions(
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    return db.execute(
        select(Subscription).where(Subscription.user_id == user.id)
    ).scalars().all()


@router.get("/{subscription_id}", response_model=SubscriptionOut)
def get_subscription(
    subscription_id: int,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    sub = db.execute(
        select(Subscription).where(
            Subscription.id == subscription_id, Subscription.user_id == user.id
        )
    ).scalar_one_or_none()
    if not sub:
        raise HTTPException(status_code=404, detail="Subscription not found")
    return sub


@router.patch("/{subscription_id}", response_model=SubscriptionOut)
def update_subscription(
    subscription_id: int,
    payload: SubscriptionUpdate,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    sub = db.execute(
        select(Subscription).where(
            Subscription.id == subscription_id, Subscription.user_id == user.id
        )
    ).scalar_one_or_none()
    if not sub:
        raise HTTPException(status_code=404, detail="Subscription not found")

    update_data = payload.model_dump(exclude_unset=True)
    if "webhook_url" in update_data:
        update_data["webhook_url"] = str(update_data["webhook_url"])
    for field, value in update_data.items():
        setattr(sub, field, value)

    db.commit()
    db.refresh(sub)

    publish_config_change("subscription_updated", sub.id)
    return sub


@router.delete("/{subscription_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_subscription(
    subscription_id: int,
    db: Session = Depends(get_db),
    user: User = Depends(get_current_user),
):
    sub = db.execute(
        select(Subscription).where(
            Subscription.id == subscription_id, Subscription.user_id == user.id
        )
    ).scalar_one_or_none()
    if not sub:
        raise HTTPException(status_code=404, detail="Subscription not found")

    db.delete(sub)
    db.commit()

    publish_config_change("subscription_deleted", subscription_id)