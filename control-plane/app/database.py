from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base

# Connection string format:
# mysql+pymysql://user:password@host:port/database
DATABASE_URL = "mysql+pymysql://root:rootpass@localhost:3308/argus_metadata"

engine = create_engine(DATABASE_URL, echo=False, pool_pre_ping=True)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()


def get_db():
    """FastAPI dependency that yields a DB session and closes it on exit."""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()