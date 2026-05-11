from fastapi import FastAPI

app = FastAPI(title="Argus Control Plane")


@app.get("/")
def root():
    return {"service": "argus-control-plane", "status": "ok"}


@app.get("/healthz")
def health():
    return {"status": "healthy"}