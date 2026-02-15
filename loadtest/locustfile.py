from locust import FastHttpUser, HttpUser, between
import random


def get_product(user):
    product_id = random.choice([1, 2])
    user.client.get(f"/products/{product_id}", name="GET /products/{productId}")


def update_product(user):
    product_id = random.choice([1, 2])
    payload = {
        "product_id": product_id,
        "sku": f"SKU-{random.randint(100, 999)}",
        "manufacturer": "Locust Inc",
        "category_id": 10,
        "weight": random.randint(0, 2000),
        "some_other_id": random.randint(1, 1000),
    }
    user.client.post(
        f"/products/{product_id}/details",
        json=payload,
        name="POST /products/{productId}/details",
    )


class ProductHttpUser(HttpUser):
    wait_time = between(0.1, 1.0)
    tasks = {get_product: 4, update_product: 1}


class ProductFastHttpUser(FastHttpUser):
    wait_time = between(0.1, 1.0)
    tasks = {get_product: 4, update_product: 1}
