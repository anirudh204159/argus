SOURCE_PAYLOAD = {
    "name": "test-source",
    "host": "argus-source-mysql",
    "port": 3306,
    "database_name": "argus_demo",
    "replication_user": "root",
    "replication_password": "rootpass",
}


def test_create_source_requires_auth(client):
    response = client.post("/sources", json=SOURCE_PAYLOAD)
    assert response.status_code == 401


def test_create_source_success(client, auth_headers):
    response = client.post("/sources", json=SOURCE_PAYLOAD, headers=auth_headers)
    assert response.status_code == 201
    data = response.json()
    assert data["name"] == SOURCE_PAYLOAD["name"]
    assert data["host"] == SOURCE_PAYLOAD["host"]
    assert "replication_password" not in data
    assert "replication_password_enc" not in data


def test_list_sources_returns_only_own(client, auth_headers):
    client.post("/sources", json=SOURCE_PAYLOAD, headers=auth_headers)

    # Create a second user
    client.post(
        "/auth/register",
        json={"email": "other@argus.dev", "password": "otherpassword"},
    )
    login = client.post(
        "/auth/login",
        json={"email": "other@argus.dev", "password": "otherpassword"},
    )
    other_headers = {"Authorization": f"Bearer {login.json()['access_token']}"}
    client.post(
        "/sources",
        json={**SOURCE_PAYLOAD, "name": "other-source"},
        headers=other_headers,
    )

    response = client.get("/sources", headers=auth_headers)
    assert response.status_code == 200
    sources = response.json()
    assert len(sources) == 1
    assert sources[0]["name"] == "test-source"


def test_get_source_not_found(client, auth_headers):
    response = client.get("/sources/9999", headers=auth_headers)
    assert response.status_code == 404


def test_cannot_get_other_users_source(client, auth_headers):
    # Other user creates a source
    client.post(
        "/auth/register",
        json={"email": "other@argus.dev", "password": "otherpassword"},
    )
    login = client.post(
        "/auth/login",
        json={"email": "other@argus.dev", "password": "otherpassword"},
    )
    other_headers = {"Authorization": f"Bearer {login.json()['access_token']}"}
    create = client.post("/sources", json=SOURCE_PAYLOAD, headers=other_headers)
    other_source_id = create.json()["id"]

    # Our user tries to access it
    response = client.get(f"/sources/{other_source_id}", headers=auth_headers)
    assert response.status_code == 404


def test_duplicate_source_name_returns_409(client, auth_headers):
    client.post("/sources", json=SOURCE_PAYLOAD, headers=auth_headers)
    response = client.post("/sources", json=SOURCE_PAYLOAD, headers=auth_headers)
    assert response.status_code == 409


def test_update_source(client, auth_headers):
    create = client.post("/sources", json=SOURCE_PAYLOAD, headers=auth_headers)
    source_id = create.json()["id"]

    response = client.patch(
        f"/sources/{source_id}",
        json={"name": "renamed-source"},
        headers=auth_headers,
    )
    assert response.status_code == 200
    assert response.json()["name"] == "renamed-source"
    assert response.json()["host"] == SOURCE_PAYLOAD["host"]  # unchanged


def test_delete_source(client, auth_headers):
    create = client.post("/sources", json=SOURCE_PAYLOAD, headers=auth_headers)
    source_id = create.json()["id"]

    response = client.delete(f"/sources/{source_id}", headers=auth_headers)
    assert response.status_code == 204

    response = client.get(f"/sources/{source_id}", headers=auth_headers)
    assert response.status_code == 404