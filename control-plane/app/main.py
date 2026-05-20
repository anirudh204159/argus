import os

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from slowapi.errors import RateLimitExceeded
from slowapi.middleware import SlowAPIMiddleware
from starlette.responses import JSONResponse

from app.rate_limit import limiter
from app.routers import auth, sources, subscriptions

app = FastAPI(title="Argus Control Plane")

# CORS — allow the frontend to call this API.
# In production, set ARGUS_CORS_ORIGINS to your real frontend URL(s).
_cors_origins = os.getenv(
    "ARGUS_CORS_ORIGINS",
    "http://localhost:5173,http://localhost:5174,http://127.0.0.1:5173",
).split(",")

app.add_middleware(
    CORSMiddleware,
    allow_origins=_cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Rate limit setup
app.state.limiter = limiter
app.add_middleware(SlowAPIMiddleware)


@app.exception_handler(RateLimitExceeded)
def rate_limit_handler(request, exc: RateLimitExceeded):
    return JSONResponse(
        status_code=429,
        content={"detail": f"Rate limit exceeded: {exc.detail}"},
    )


app.include_router(auth.router)
app.include_router(sources.router)
app.include_router(subscriptions.router)


@app.get("/")
def root():
    return {"service": "argus-control-plane", "status": "ok"}


@app.get("/healthz")
def health():
    return {"status": "healthy"}