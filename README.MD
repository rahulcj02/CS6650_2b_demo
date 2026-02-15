## CS6650 Homework 5 - Product API + Terraform ECS/ECR

This repository implements the **Product API** portion of the provided OpenAPI spec using Go + Gin, containers it with Docker, and deploys to AWS ECS/Fargate using Terraform.

Piazza tasks are intentionally excluded here (per your request).

## Repository Layout
- `src/main.go`: Product API server implementation
- `src/main_test.go`: API tests for expected response codes
- `src/Dockerfile`: container image for the API server
- `terraform/`: AWS infrastructure as code (ECR + ECS + networking + logging)
- `loadtest/locustfile.py`: Locust load tests (`HttpUser` and `FastHttpUser`)

## Implemented Product Endpoints
- `GET /products/{productId}`
  - `200` product found
  - `404` product not found
- `POST /products/{productId}/details`
  - `204` product details updated
  - `400` invalid payload/path mismatch
  - `404` product not found

Data is stored in-memory via a Go map protected by `sync.RWMutex`.

## Local Run
```bash
cd src
go run .
```
Server runs at `http://localhost:8080`.

## Local Tests
```bash
cd src
go test ./...
```

## Curl Examples (Response Codes)
Run server locally first.

`GET 200`
```bash
curl -i http://localhost:8080/products/1
```

`GET 404`
```bash
curl -i http://localhost:8080/products/999
```

`POST 204`
```bash
curl -i -X POST http://localhost:8080/products/1/details \
  -H 'Content-Type: application/json' \
  -d '{
    "product_id": 1,
    "sku": "SKU-123",
    "manufacturer": "Acme Corporation",
    "category_id": 100,
    "weight": 1200,
    "some_other_id": 501
  }'
```

`POST 400` (missing required `sku`)
```bash
curl -i -X POST http://localhost:8080/products/1/details \
  -H 'Content-Type: application/json' \
  -d '{
    "product_id": 1,
    "manufacturer": "Acme Corporation",
    "category_id": 100,
    "weight": 1200,
    "some_other_id": 501
  }'
```

`POST 404`
```bash
curl -i -X POST http://localhost:8080/products/999/details \
  -H 'Content-Type: application/json' \
  -d '{
    "product_id": 999,
    "sku": "SKU-999",
    "manufacturer": "Acme Corporation",
    "category_id": 100,
    "weight": 1200,
    "some_other_id": 501
  }'
```

`500` is reserved for unexpected server-side failures and is not intentionally forced by normal test input.

## Docker
```bash
cd src
docker build -t product-api:local .
docker run --rm -p 8080:8080 product-api:local
```

## AWS Deployment (Terraform)
### 1) Configure AWS credentials (Learner Lab)
```bash
aws configure
aws configure set aws_session_token <YOUR_TEMP_SESSION_TOKEN>
```

### 2) Deploy infrastructure + app
```bash
cd terraform
terraform init
terraform apply -auto-approve
```

### 3) Get public IP of running ECS task
Run from `terraform/`:
```bash
aws ec2 describe-network-interfaces \
--network-interface-ids $(
  aws ecs describe-tasks \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --tasks $(
    aws ecs list-tasks \
    --cluster $(terraform output -raw ecs_cluster_name) \
    --service-name $(terraform output -raw ecs_service_name) \
    --query 'taskArns[0]' --output text
  ) \
  --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
  --output text
) \
--query 'NetworkInterfaces[0].Association.PublicIp' \
--output text
```

### 4) Call deployed API
```bash
curl -i http://<PUBLIC-IP>:8080/products/1
```

### 5) Check logs in CloudWatch
Log group name pattern: `/ecs/CS6650L2`

### 6) Clean up
```bash
cd terraform
terraform destroy -auto-approve
```

## Load Testing (Locust)
Install Locust if needed:
```bash
python3 -m pip install locust
```

Run with regular HTTP client:
```bash
locust -f loadtest/locustfile.py ProductHttpUser --host http://<PUBLIC-IP>:8080
```

Run with fast HTTP client:
```bash
locust -f loadtest/locustfile.py ProductFastHttpUser --host http://<PUBLIC-IP>:8080
```

Suggested settings for comparable runs:
- Users: `100`
- Spawn rate: `10`
- Duration: `2m`

Headless mode (saves CSVs for report):
```bash
source .venv/bin/activate
locust -f loadtest/locustfile.py ProductHttpUser --host http://<PUBLIC-IP>:8080 --headless -u 50 -r 10 -t 30s --only-summary --exit-code-on-error 0 --csv=loadtest/results/httpuser
locust -f loadtest/locustfile.py ProductFastHttpUser --host http://<PUBLIC-IP>:8080 --headless -u 50 -r 10 -t 30s --only-summary --exit-code-on-error 0 --csv=loadtest/results/fasthttpuser
```

Sample measured results (Feb 15, 2026):
- `ProductHttpUser`: `~73.29 req/s`, `avg ~95 ms`, `0% failures`
- `ProductFastHttpUser`: `~73.44 req/s`, `avg ~94 ms`, `0% failures`

Why results are close:
- This workload is network/server-bound at low payload sizes.
- When backend/network dominates latency, `FastHttpUser` and `HttpUser` can look similar.
- Bigger differences usually appear at higher concurrency, larger payloads, or slower client-side processing.

## Report Screenshot Checklist
Store these screenshots for your report submission:
1. Terminal showing `go test ./...` passing.
2. Local `curl` responses proving `200`, `204`, `400`, `404`.
3. Docker build and running container output.
4. `terraform apply -auto-approve` success summary.
5. AWS ECS service/task running in AWS Console.
6. CloudWatch logs showing API requests.
7. Locust results for `ProductHttpUser` (Statistics + Charts).
8. Locust results for `ProductFastHttpUser` (Statistics + Charts).
9. `terraform destroy -auto-approve` success summary.

## Short Answers: Questions To Explore
### 1) Scalable backend design for the full `api.yaml`
- Split into services by domain: Product, Cart, Warehouse, Payment.
- Use API Gateway + auth in front of stateless services on ECS/Kubernetes.
- Store products/inventory/orders in durable databases (e.g., Aurora/DynamoDB).
- Use async messaging (SQS/SNS/Kafka) for checkout, reservation, shipping workflows.
- Add distributed tracing, centralized logs, and metrics for reliability/observability.
- Enforce idempotency and retries for payment and checkout operations.

### 2) Terraform declarative vs imperative
- Declarative means you define the **desired end state** (what resources should exist).
- Imperative means writing exact step-by-step commands (how to create each resource).
- Terraform computes a plan (diff) from current to desired state and applies it safely.
- This reduces manual drift, makes infra reproducible, and supports versioned changes.

## Notes on Security / Repo Hygiene
Sensitive and large generated files are ignored via `.gitignore` (`*.tfstate*`, `.terraform/`, `.env*`, keys).
