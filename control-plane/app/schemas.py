from datetime import datetime
from typing import Literal
from pydantic import BaseModel, EmailStr, Field


# ---------- Auth ----------

class UserRegister(BaseModel):
    """Request body for POST /auth/register"""
    email: EmailStr
    password: str = Field(min_length=8, max_length=128)


class UserLogin(BaseModel):
    """Request body for POST /auth/login"""
    email: EmailStr
    password: str


class UserOut(BaseModel):
    """User data returned to clients (no password hash)"""
    id: int
    email: EmailStr
    created_at: datetime
    last_login_at: datetime | None = None

    class Config:
        from_attributes = True  # lets Pydantic build this from a SQLAlchemy model


class Token(BaseModel):
    """Response body for successful login"""
    access_token: str
    token_type: str = "bearer"


# ---------- Sources ----------

class SourceCreate(BaseModel):
    """Request body for POST /sources"""
    name: str = Field(min_length=1, max_length=100)
    host: str = Field(min_length=1, max_length=255)
    port: int = Field(default=3306, ge=1, le=65535)
    database_name: str = Field(min_length=1, max_length=100)
    replication_user: str = Field(min_length=1, max_length=100)
    replication_password: str = Field(min_length=1, max_length=255)


class SourceUpdate(BaseModel):
    """Request body for PATCH /sources/{id}. All fields optional."""
    name: str | None = Field(default=None, min_length=1, max_length=100)
    host: str | None = Field(default=None, min_length=1, max_length=255)
    port: int | None = Field(default=None, ge=1, le=65535)
    database_name: str | None = Field(default=None, min_length=1, max_length=100)
    replication_user: str | None = Field(default=None, min_length=1, max_length=100)
    replication_password: str | None = Field(default=None, min_length=1, max_length=255)


class SourceOut(BaseModel):
    """Source data returned to clients (no replication password)."""
    id: int
    name: str
    host: str
    port: int
    database_name: str
    replication_user: str
    status: str
    last_error: str | None = None
    created_at: datetime

    class Config:
        from_attributes = True

# ---------- Subscriptions ----------

# Allowed DB operations a subscription can listen for.
Operation = Literal["INSERT", "UPDATE", "DELETE"]


class SubscriptionCreate(BaseModel):
    """Request body for POST /subscriptions"""
    source_id: int
    name: str = Field(min_length=1, max_length=100)
    tables: list[str] = Field(min_length=1)
    operations: list[Operation] = Field(min_length=1)
    webhook_url: str = Field(min_length=1, max_length=2000)
    retry_max: int = Field(default=5, ge=0, le=20)


class SubscriptionUpdate(BaseModel):
    """Request body for PATCH /subscriptions/{id}. All fields optional."""
    name: str | None = Field(default=None, min_length=1, max_length=100)
    tables: list[str] | None = Field(default=None, min_length=1)
    operations: list[Operation] | None = Field(default=None, min_length=1)
    webhook_url: str | None = Field(default=None, min_length=1, max_length=2000)
    retry_max: int | None = Field(default=None, ge=0, le=20)
    active: bool | None = None


class SubscriptionOut(BaseModel):
    """Subscription data returned to clients."""
    id: int
    source_id: int
    name: str
    tables: list[str]
    operations: list[str]
    webhook_url: str
    retry_max: int
    active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class SubscriptionCreated(SubscriptionOut):
    """Returned on initial creation — includes the HMAC secret (only shown once)."""
    hmac_secret: str