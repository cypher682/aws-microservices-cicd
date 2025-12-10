import pytest
from fastapi.testclient import TestClient
from src.main import app

client = TestClient(app)

def test_health_endpoint():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "healthy", "service": "user-service"}

def test_create_user():
    user_data = {
        "email": "test@example.com",
        "name": "Test User",
        "age": 25
    }
    response = client.post("/", json=user_data)
    assert response.status_code == 201
    data = response.json()
    assert "userId" in data
    assert data["email"] == user_data["email"]
    assert data["name"] == user_data["name"]
