import time

from app.llm import router


def test_extract_retry_after_seconds_from_groq_style_message():
    text = "Rate limit reached. Please try again in 1.88s. Need more tokens?"
    assert router._extract_retry_after_seconds(text) == 1


def test_extract_retry_after_seconds_from_retry_in_phrasing():
    assert router._extract_retry_after_seconds("please retry in 3.5s") == 3
    assert router._extract_retry_after_seconds("retry in 0.2s") == 1


def test_extract_retry_after_seconds_from_retry_after_phrasing():
    assert router._extract_retry_after_seconds("Retry-After: 42s") == 42


def test_extract_retry_after_seconds_returns_none_when_absent():
    assert router._extract_retry_after_seconds("some unrelated error") is None


def test_cooldown_for_error_billing():
    seconds, reason = router._cooldown_for_error(
        "groq", Exception("402 Payment Required")
    )
    assert seconds == 3600
    assert "billing" in reason


def test_cooldown_for_error_auth():
    seconds, reason = router._cooldown_for_error(
        "gemini", Exception("401 Unauthorized")
    )
    assert seconds == 3600
    assert "authentication" in reason


def test_cooldown_for_error_rate_limit_uses_retry_after_when_present():
    exc = Exception("429 rate limit exceeded, retry in 5s")
    seconds, reason = router._cooldown_for_error("groq", exc)
    assert seconds == 5
    assert reason == "rate limited"


def test_cooldown_for_error_rate_limit_falls_back_to_60s():
    exc = Exception("429 rate limit exceeded")
    seconds, reason = router._cooldown_for_error("groq", exc)
    assert seconds == 60


def test_cooldown_for_error_transient_server_error():
    seconds, reason = router._cooldown_for_error(
        "cerebras", Exception("503 Service Unavailable")
    )
    assert seconds == 15
    assert "transient" in reason


def test_cooldown_for_error_timeout():
    seconds, reason = router._cooldown_for_error(
        "mistral", Exception("request timed out")
    )
    assert seconds == 15
    assert "timeout" in reason


def test_cooldown_for_error_unknown_defaults_to_10s():
    seconds, reason = router._cooldown_for_error(
        "ollama", Exception("connection reset")
    )
    assert seconds == 10


def test_record_failure_then_success_resets_availability():
    provider = "test-provider-" + str(time.monotonic())

    assert router._is_available(provider) is True

    router._record_failure(provider, Exception("500 internal error"))
    assert router._is_available(provider) is False
    assert router._provider_failure_counts[provider] == 1

    router._record_success(provider)
    assert router._is_available(provider) is True
    assert router._provider_failure_counts[provider] == 0


def test_record_failure_never_shortens_an_existing_longer_cooldown():
    provider = "test-provider-monotonic-" + str(time.monotonic())

    router._record_failure(provider, Exception("402 payment required"))
    long_cooldown = router._provider_cooldown_until[provider]

    router._record_failure(provider, Exception("connection reset"))
    assert router._provider_cooldown_until[provider] == long_cooldown
