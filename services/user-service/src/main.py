from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, EmailStr
from typing import Optional
import boto3
from datetime import datetime
import uuid
import os
from prometheus_client import Counter, Histogram, generate_latest, CONTENT_TYPE_LATEST
from fastapi.responses import Response

app = FastAPI(title="User Service")

dynamodb = boto3.resource('dynamodb', region_name=os.getenv('AWS_REGION', 'us-east-1'))
table = dynamodb.Table(os.getenv('DYNAMODB_USERS_TABLE', 'aws-microservices-cicd-users'))

request_count = Counter('user_service_requests_total', 'Total requests', ['method', 'endpoint', 'status'])
request_duration = Histogram('user_service_request_duration_seconds', 'Request duration', ['method', 'endpoint'])

class User(BaseModel):
    email: EmailStr
    name: str
    age: Optional[int] = None

class UserResponse(BaseModel):
    userId: str
    email: str
    name: str
    age: Optional[int]
    createdAt: str
    updatedAt: str

@app.get("/health")
def health():
    return {"status": "healthy", "service": "user-service"}

@app.get("/metrics")
def metrics():
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)

@app.post("/", response_model=UserResponse, status_code=201)
async def create_user(user: User):
    with request_duration.labels(method='POST', endpoint='/').time():
        try:
            user_id = str(uuid.uuid4())
            timestamp = datetime.utcnow().isoformat()
            
            item = {
                'userId': user_id,
                'email': user.email,
                'name': user.name,
                'age': user.age,
                'createdAt': timestamp,
                'updatedAt': timestamp
            }
            
            table.put_item(
                Item=item,
                ConditionExpression='attribute_not_exists(userId)'
            )
            
            request_count.labels(method='POST', endpoint='/', status=201).inc()
            return item
        except Exception as e:
            request_count.labels(method='POST', endpoint='/', status=500).inc()
            raise HTTPException(status_code=500, detail=str(e))

@app.get("/{user_id}", response_model=UserResponse)
async def get_user(user_id: str):
    with request_duration.labels(method='GET', endpoint='/{user_id}').time():
        try:
            response = table.get_item(Key={'userId': user_id})
            
            if 'Item' not in response:
                request_count.labels(method='GET', endpoint='/{user_id}', status=404).inc()
                raise HTTPException(status_code=404, detail="User not found")
            
            request_count.labels(method='GET', endpoint='/{user_id}', status=200).inc()
            return response['Item']
        except HTTPException:
            raise
        except Exception as e:
            request_count.labels(method='GET', endpoint='/{user_id}', status=500).inc()
            raise HTTPException(status_code=500, detail=str(e))

@app.get("/")
async def list_users(limit: int = 10):
    with request_duration.labels(method='GET', endpoint='/').time():
        try:
            response = table.scan(Limit=limit)
            request_count.labels(method='GET', endpoint='/', status=200).inc()
            return {"users": response.get('Items', []), "count": len(response.get('Items', []))}
        except Exception as e:
            request_count.labels(method='GET', endpoint='/', status=500).inc()
            raise HTTPException(status_code=500, detail=str(e))

@app.put("/{user_id}", response_model=UserResponse)
async def update_user(user_id: str, user: User):
    with request_duration.labels(method='PUT', endpoint='/{user_id}').time():
        try:
            timestamp = datetime.utcnow().isoformat()
            
            response = table.update_item(
                Key={'userId': user_id},
                UpdateExpression='SET #name = :name, email = :email, age = :age, updatedAt = :updated',
                ExpressionAttributeNames={'#name': 'name'},
                ExpressionAttributeValues={
                    ':name': user.name,
                    ':email': user.email,
                    ':age': user.age,
                    ':updated': timestamp
                },
                ConditionExpression='attribute_exists(userId)',
                ReturnValues='ALL_NEW'
            )
            
            request_count.labels(method='PUT', endpoint='/{user_id}', status=200).inc()
            return response['Attributes']
        except dynamodb.meta.client.exceptions.ConditionalCheckFailedException:
            request_count.labels(method='PUT', endpoint='/{user_id}', status=404).inc()
            raise HTTPException(status_code=404, detail="User not found")
        except Exception as e:
            request_count.labels(method='PUT', endpoint='/{user_id}', status=500).inc()
            raise HTTPException(status_code=500, detail=str(e))

@app.delete("/{user_id}")
async def delete_user(user_id: str):
    with request_duration.labels(method='DELETE', endpoint='/{user_id}').time():
        try:
            table.delete_item(
                Key={'userId': user_id},
                ConditionExpression='attribute_exists(userId)'
            )
            request_count.labels(method='DELETE', endpoint='/{user_id}', status=204).inc()
            return Response(status_code=204)
        except dynamodb.meta.client.exceptions.ConditionalCheckFailedException:
            request_count.labels(method='DELETE', endpoint='/{user_id}', status=404).inc()
            raise HTTPException(status_code=404, detail="User not found")
        except Exception as e:
            request_count.labels(method='DELETE', endpoint='/{user_id}', status=500).inc()
            raise HTTPException(status_code=500, detail=str(e))
