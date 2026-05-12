from datetime import datetime
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