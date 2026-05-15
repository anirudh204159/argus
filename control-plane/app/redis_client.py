import redis

REDIS_URL = "redis://localhost:6379/0"
CONFIG_CHANNEL = "argus:config"

_client = redis.Redis.from_url(REDIS_URL, decode_responses=True)


def publish_config_change(event_type: str, subscription_id: int | None = None) -> None:
    """
    Notify subscribers that the configuration has changed.

    event_type: "subscription_created" / "subscription_updated" /
                "subscription_deleted" / "source_updated" / etc.
    subscription_id: optional, for future targeted reloads
    """
    try:
        message = f"{event_type}:{subscription_id or ''}"
        _client.publish(CONFIG_CHANNEL, message)
    except redis.RedisError as e:
        # We don't want a Redis outage to break the API.
        # Worst case: worker still polls every 30s as fallback.
        print(f"redis publish failed (non-fatal): {e}")