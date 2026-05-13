def test_register_new_user(client):
    response = client.post(
        "/auth/register",
        json={"email": "new@argus.dev", "password": "supersecret"},
    )
    assert response.status_code == 201
    data = response.json()
    assert data["email"] == "new@argus.dev"
    assert "id" in data
    assert "password_hash" not in data
    assert "password" not in data


def test_register_duplicate_email_returns_409(client):
    client.post(
        "/auth/register",
        json={"email": "dup@argus.dev", "password": "supersecret"},
    )
    response = client.post(
        "/auth/register",
        json={"email": "dup@argus.dev", "password": "anotherpassword"},
    )
    assert response.status_code == 409


def test_register_short_password_returns_422(client):
    response = client.post(
        "/auth/register",
        json={"email": "short@argus.dev", "password": "abc"},
    )
    assert response.status_code == 422


def test_register_invalid_email_returns_422(client):
    response = client.post(
        "/auth/register",
        json={"email": "not-an-email", "password": "supersecret"},
    )
    assert response.status_code == 422


def test_login_success_returns_token(client):
    client.post(
        "/auth/register",
        json={"email": "login@argus.dev", "password": "supersecret"},
    )
    response = client.post(
        "/auth/login",
        json={"email": "login@argus.dev", "password": "supersecret"},
    )
    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert data["token_type"] == "bearer"


def test_login_wrong_password_returns_401(client):
    client.post(
        "/auth/register",
        json={"email": "wrong@argus.dev", "password": "supersecret"},
    )
    response = client.post(
        "/auth/login",
        json={"email": "wrong@argus.dev", "password": "wrongpass"},
    )
    assert response.status_code == 401


def test_login_unknown_email_returns_401(client):
    response = client.post(
        "/auth/login",
        json={"email": "ghost@argus.dev", "password": "supersecret"},
    )
    assert response.status_code == 401