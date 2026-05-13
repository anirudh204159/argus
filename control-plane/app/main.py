from fastapi import FastAPI

from app.routers import auth, sources, subscriptions

app = FastAPI(title="Argus Control Plane")

app.include_router(auth.router)
app.include_router(sources.router)
app.include_router(subscriptions.router)


@app.get("/")
def root():
    return {"service": "argus-control-plane", "status": "ok"}


@app.get("/healthz")
def health():
    return {"status": "healthy"}