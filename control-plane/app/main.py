from fastapi import FastAPI

from app.routers import auth

app = FastAPI(title="Argus Control Plane")

app.include_router(auth.router)


@app.get("/")
def root():
    return {"service": "argus-control-plane", "status": "ok"}


@app.get("/healthz")
def health():
    return {"status": "healthy"}