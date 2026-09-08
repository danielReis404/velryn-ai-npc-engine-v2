from fastapi.testclient import TestClient

from app import main

client = TestClient(main.app)


def test_health_endpoint():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_decide_falls_back_to_wait_when_the_graph_raises(
    monkeypatch, base_perception_kwargs
):
    async def _raise(*args, **kwargs):
        raise RuntimeError("every provider failed")

    monkeypatch.setattr(main.compiled_graph, "ainvoke", _raise)

    response = client.post("/decide", json=base_perception_kwargs)

    assert response.status_code == 200
    body = response.json()
    assert body["action"] == "WAIT"


def test_decide_rejects_a_malformed_payload_with_422():
    response = client.post("/decide", json={"name": "Selene"})
    assert response.status_code == 422
