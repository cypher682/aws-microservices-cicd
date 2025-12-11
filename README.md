# AWS Microservices CI/CD

Production-ready microservices architecture with complete CI/CD pipeline, infrastructure as code, and monitoring stack.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       AWS Cloud (us-east-1)                  │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │              VPC (10.0.0.0/16)                     │    │
│  │                                                     │    │
│  │  ┌──────────────────────────────────────────┐     │    │
│  │  │   EC2 t2.micro (Public Subnet)           │     │    │
│  │  │                                           │     │    │
│  │  │  ┌─────────────────────────────────┐    │     │    │
│  │  │  │  Nginx (SSL/TLS)                │    │     │    │
│  │  │  └──────────┬──────────────────────┘    │     │    │
│  │  │             │                             │     │    │
│  │  │  ┌──────────▼──────────────────────┐    │     │    │
│  │  │  │  API Gateway (Node.js)          │    │     │    │
│  │  │  │  Port: 3000                      │    │     │    │
│  │  │  └──────┬─────────────┬────────────┘    │     │    │
│  │  │         │             │                  │     │    │
│  │  │  ┌──────▼─────┐  ┌───▼──────────┐      │     │    │
│  │  │  │ User Service│  │Product Service│      │     │    │
│  │  │  │  (Python)   │  │    (Go)       │      │     │    │
│  │  │  │  Port: 8000 │  │  Port: 8080   │      │     │    │
│  │  │  └──────┬──────┘  └───┬──────────┘      │     │    │
│  │  │         │              │                  │     │    │
│  │  │  ┌──────▼──────────────▼─────────┐      │     │    │
│  │  │  │  Prometheus (Port: 9090)      │      │     │    │
│  │  │  └───────────────────────────────┘      │     │    │
│  │  │                                           │     │    │
│  │  │  ┌──────────────────────────────┐       │     │    │
│  │  │  │  Grafana (Port: 3001)        │       │     │    │
│  │  │  └──────────────────────────────┘       │     │    │
│  │  └───────────────────────────────────────────┘    │    │
│  │                                                     │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────────┐        ┌──────────────────┐           │
│  │  DynamoDB       │        │  Route53         │           │
│  │  - Users        │        │  api.yourname│          │
│  │  - Products     │        └──────────────────┘           │
│  └─────────────────┘                                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘

GitHub Actions → Build → Trivy Scan → Docker Hub → Deploy to EC2
```

## Features

- **Polyglot Microservices**: Node.js, Python, Go
- **Infrastructure as Code**: Terraform for AWS resources
- **CI/CD Pipeline**: GitHub Actions with automated testing and deployment
- **Security**: Container scanning, SSL/TLS, IAM least-privilege
- **Monitoring**: Prometheus + Grafana with custom dashboards
- **Configuration Management**: Ansible for EC2 setup
- **Cost Optimized**: Free tier eligible ($0.50-2/month)

## Tech Stack

- **API Gateway**: Node.js + Express
- **User Service**: Python + FastAPI
- **Product Service**: Go + Gin
- **Database**: AWS DynamoDB
- **Container**: Docker + Docker Compose
- **Orchestration**: Docker Compose on EC2
- **Reverse Proxy**: Nginx with Let's Encrypt SSL
- **Monitoring**: Prometheus + Grafana
- **IaC**: Terraform
- **CM**: Ansible
- **CI/CD**: GitHub Actions
- **Registry**: Docker Hub

## Prerequisites

### Windows Host
- Git Bash
- AWS CLI configured
- Terraform
- Docker Desktop
- SSH client

### WSL (Ubuntu 24.04)
- Ansible
- Python 3
- SSH client

## Quick Start

### 1. Clone Repository

```bash
cd C:/Users/<YOUR_USERNAME>
git clone https://github.com/<YOUR_USERNAME>/aws-microservices-cicd.git
cd aws-microservices-cicd
```

### 2. Generate SSH Key

```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/aws-microservices-key -N ""
```

### 3. Configure GitHub Secrets

Add these secrets in GitHub repository settings:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `DOCKER_USERNAME`: your-name
- `DOCKER_PASSWORD`
- `EC2_SSH_PRIVATE_KEY`
- `API_KEY`: Generate with `openssl rand -hex 32`

### 4. Deploy Infrastructure

```bash
cd terraform

terraform init

terraform plan

terraform apply

terraform output nameservers
```

### 5. Update GoDaddy DNS

Update nameservers in GoDaddy to Route53 nameservers from output above. Wait 10-60 minutes for DNS propagation.

### 6. Setup EC2 (From WSL)

```bash
cd /mnt/c/Users/<YOUR_USERNAME>/aws-microservices-cicd

ansible-playbook -i ansible/inventory/hosts ansible/playbooks/setup.yml
```

### 7. Deploy Application

Push to main branch to trigger automatic deployment:

```bash
git add .
git commit -m "Initial deployment"
git push origin main
```

Or manually deploy via Ansible:

```bash
ansible-playbook -i ansible/inventory/hosts ansible/playbooks/deploy.yml
```

## API Usage

### Get API Key

```bash
aws ssm get-parameter --name /aws-microservices-cicd/api-key --with-decryption --query 'Parameter.Value' --output text
```

### Test Endpoints

```bash
export API_KEY="your-api-key"

curl -H "X-API-Key: $API_KEY" https://api.domain/health

curl -X POST https://api.domain/users \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "name": "Test User", "age": 25}'

curl -H "X-API-Key: $API_KEY" https://api.domain/users

curl -X POST https://api.domain/products \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "Product A", "description": "Test product", "price": 99.99, "category": "Electronics", "stock": 100}'

curl -H "X-API-Key: $API_KEY" https://api.domain/products
```

## Monitoring

Access Grafana at https://api.yourdomain:3001
- Username: admin
- Password: admin

Access Prometheus at https://api.domain:9090

## Project Structure

```
aws-microservices-cicd/
├── .github/workflows/         # CI/CD pipelines
├── terraform/                 # Infrastructure code
├── services/                  # Microservices
│   ├── api-gateway/          # Node.js API Gateway
│   ├── user-service/         # Python User Service
│   └── product-service/      # Go Product Service
├── ansible/                   # Configuration management
├── monitoring/                # Prometheus + Grafana configs
├── nginx/                     # Nginx configuration
└── docker-compose.yml         # Container orchestration
```

## Cost Breakdown

- Route53 Hosted Zone: $0.50/month
- EC2 t2.micro: Free tier (750 hours/month)
- DynamoDB: Free tier (25GB, 25 RCU/WCU)
- Data Transfer: $0-1/month

**Total: $0.50-2.00/month**

## Security Features

- SSL/TLS encryption (Let's Encrypt)
- Container vulnerability scanning (Trivy)
- IAM least-privilege policies
- Security groups with minimal ports
- API key authentication
- Secrets stored in Parameter Store

## Troubleshooting

### SSH Connection Failed
```bash
chmod 600 ~/.ssh/aws-microservices-key
ssh -i ~/.ssh/aws-microservices-key ubuntu@<EC2_IP>
```

### SSL Certificate Issues
```bash
ssh -i ~/.ssh/aws-microservices-key ubuntu@<EC2_IP>
sudo certbot --nginx -d "api.yourdomain"
```

### Container Issues
```bash
ssh -i ~/.ssh/aws-microservices-key ubuntu@<EC2_IP>
cd /home/ubuntu/app
docker-compose logs
docker-compose restart
```

## License

MIT

## Author

Built for learning and demonstration purposes.
