SOURCE_PAYLOAD = {
    "name": "test-source",
    "host": "argus-source-mysql",
    "port": 3306,
    "database_name": "argus_demo",
    "replication_user": "root",
    "replication_password": "rootpass",
}


def _create_source(client, headers):
    """Helper: create a source and return its id."""
    response = client.post("/sources", json=SOURCE_PAYLOAD, headers=headers)
    return response.json()["id"]


def _subscription_payload(source_id):
    return {
        "source_id": source_id,
        "name": "orders-webhook",
        "tables": ["orders"],
        "operations": ["INSERT", "UPDATE"],
        "webhook_url": "https://webhook.site/test",
        "retry_max": 3,
    }


def test_create_subscription_requires_auth(client):
    response = client.post("/subscriptions", json=_subscription_payload(1))
    assert response.status_code == 401


def test_create_subscription_success_returns_hmac_secret(client, auth_headers):
    source_id = _create_source(client, auth_headers)
    response = client.post(
        "/subscriptions",
        json=_subscription_payload(source_id),
        headers=auth_headers,
    )
    assert response.status_code == 201
    data = response.json()
    assert data["name"] == "orders-webhook"
    assert "hmac_secret" in data
    assert len(data["hmac_secret"]) > 20  # token_urlsafe(32) is at least ~43 chars


def test_list_subscriptions_does_not_return_hmac_secret(client, auth_headers):
    source_id = _create_source(client, auth_headers)
    client.post(
        "/subscriptions",
        json=_subscription_payload(source_id),
        headers=auth_headers,
    )

    response = client.get("/subscriptions", headers=auth_headers)
    assert response.status_code == 200
    subscriptions = response.json()
    assert len(subscriptions) == 1
    assert "hmac_secret" not in subscriptions[0]


def test_cannot_create_subscription_for_other_users_source(client, auth_headers):
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
    other_source_id = _create_source(client, other_headers)

    # Our user tries to subscribe to it
    response = client.post(
        "/subscriptions",
        json=_subscription_payload(other_source_id),
        headers=auth_headers,
    )
    assert response.status_code == 404


def test_invalid_operation_returns_422(client, auth_headers):
    source_id = _create_source(client, auth_headers)
    payload = _subscription_payload(source_id)
    payload["operations"] = ["INSERT", "TRUNCATE"]

    response = client.post("/subscriptions", json=payload, headers=auth_headers)
    assert response.status_code == 422


def test_empty_tables_returns_422(client, auth_headers):
    source_id = _create_source(client, auth_headers)
    payload = _subscription_payload(source_id)
    payload["tables"] = []

    response = client.post("/subscriptions", json=payload, headers=auth_headers)
    assert response.status_code == 422


def test_pause_subscription(client, auth_headers):
    source_id = _create_source(client, auth_headers)
    create = client.post(
        "/subscriptions",
        json=_subscription_payload(source_id),
        headers=auth_headers,
    )
    sub_id = create.json()["id"]

    response = client.patch(
        f"/subscriptions/{sub_id}",
        json={"active": False},
        headers=auth_headers,
    )
    assert response.status_code == 200
    assert response.json()["active"] is False


def test_delete_subscription(client, auth_headers):
    source_id = _create_source(client, auth_headers)
    create = client.post(
        "/subscriptions",
        json=_subscription_payload(source_id),
        headers=auth_headers,
    )
    sub_id = create.json()["id"]

    response = client.delete(f"/subscriptions/{sub_id}", headers=auth_headers)
    assert response.status_code == 204

    response = client.get(f"/subscriptions/{sub_id}", headers=auth_headers)
    assert response.status_code == 404


def test_delete_source_cascades_to_subscriptions(client, auth_headers):
    source_id = _create_source(client, auth_headers)
    client.post(
        "/subscriptions",
        json=_subscription_payload(source_id),
        headers=auth_headers,
    )

    client.delete(f"/sources/{source_id}", headers=auth_headers)

    response = client.get("/subscriptions", headers=auth_headers)
    assert response.status_code == 200
    assert response.json() == []