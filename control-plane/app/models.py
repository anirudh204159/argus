from datetime import datetime
from sqlalchemy import (
    Column, BigInteger, Integer, String, Text, Boolean,
    DateTime, Enum, JSON, ForeignKey, LargeBinary, UniqueConstraint, Index,
)
from sqlalchemy.orm import relationship
from app.database import Base


class User(Base):
    __tablename__ = "users"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    email = Column(String(255), nullable=False, unique=True, index=True)
    password_hash = Column(String(255), nullable=False)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)
    last_login_at = Column(DateTime, nullable=True)

    sources = relationship("Source", back_populates="user", cascade="all, delete-orphan")
    subscriptions = relationship("Subscription", back_populates="user", cascade="all, delete-orphan")


class Source(Base):
    __tablename__ = "sources"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(BigInteger, ForeignKey("users.id"), nullable=False)
    name = Column(String(100), nullable=False)
    host = Column(String(255), nullable=False)
    port = Column(Integer, nullable=False, default=3306)
    database_name = Column(String(100), nullable=False)
    replication_user = Column(String(100), nullable=False)
    replication_password_enc = Column(LargeBinary(512), nullable=False)
    status = Column(
        Enum("disconnected", "connected", "error", name="source_status"),
        nullable=False,
        default="disconnected",
    )
    last_error = Column(Text, nullable=True)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)

    user = relationship("User", back_populates="sources")
    subscriptions = relationship("Subscription", back_populates="source", cascade="all, delete-orphan")
    checkpoint = relationship("SourceCheckpoint", back_populates="source", uselist=False, cascade="all, delete-orphan")

    __table_args__ = (UniqueConstraint("user_id", "name", name="uq_source_user_name"),)


class Subscription(Base):
    __tablename__ = "subscriptions"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(BigInteger, ForeignKey("users.id"), nullable=False)
    source_id = Column(BigInteger, ForeignKey("sources.id"), nullable=False)
    name = Column(String(100), nullable=False)
    tables = Column(JSON, nullable=False)
    operations = Column(JSON, nullable=False)
    webhook_url = Column(String(2000), nullable=False)
    hmac_secret = Column(String(255), nullable=False)
    retry_max = Column(Integer, nullable=False, default=5)
    active = Column(Boolean, nullable=False, default=True)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)

    user = relationship("User", back_populates="subscriptions")
    source = relationship("Source", back_populates="subscriptions")
    deliveries = relationship("DeliveryLog", back_populates="subscription", cascade="all, delete-orphan")
    dead_letters = relationship("DeadLetterEvent", back_populates="subscription", cascade="all, delete-orphan")

    __table_args__ = (UniqueConstraint("user_id", "name", name="uq_subscription_user_name"),)


class SourceCheckpoint(Base):
    __tablename__ = "source_checkpoints"

    source_id = Column(BigInteger, ForeignKey("sources.id", ondelete="CASCADE"), primary_key=True)
    binlog_file = Column(String(255), nullable=False)
    binlog_pos = Column(BigInteger, nullable=False)
    updated_at = Column(DateTime, nullable=False, default=datetime.utcnow, onupdate=datetime.utcnow)

    source = relationship("Source", back_populates="checkpoint")


class DeliveryLog(Base):
    __tablename__ = "delivery_log"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    subscription_id = Column(BigInteger, ForeignKey("subscriptions.id", ondelete="CASCADE"), nullable=False)
    event_id = Column(String(64), nullable=False)
    attempt = Column(Integer, nullable=False)
    status = Column(Enum("success", "failure", name="delivery_status"), nullable=False)
    http_status = Column(Integer, nullable=True)
    latency_ms = Column(Integer, nullable=False)
    error_message = Column(Text, nullable=True)
    delivered_at = Column(DateTime, nullable=False, default=datetime.utcnow)

    subscription = relationship("Subscription", back_populates="deliveries")

    __table_args__ = (Index("idx_subscription_time", "subscription_id", "delivered_at"),)


class DeadLetterEvent(Base):
    __tablename__ = "dead_letter_events"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    subscription_id = Column(BigInteger, ForeignKey("subscriptions.id", ondelete="CASCADE"), nullable=False)
    event_id = Column(String(64), nullable=False)
    payload = Column(JSON, nullable=False)
    attempts = Column(Integer, nullable=False)
    last_error = Column(Text, nullable=False)
    failed_at = Column(DateTime, nullable=False, default=datetime.utcnow)
    replayed_at = Column(DateTime, nullable=True)

    subscription = relationship("Subscription", back_populates="dead_letters")

    __table_args__ = (Index("idx_subscription_failed", "subscription_id", "failed_at"),)