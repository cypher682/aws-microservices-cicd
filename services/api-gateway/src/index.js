const express = require('express');
const axios = require('axios');
const morgan = require('morgan');
const promClient = require('prom-client');
const { SSMClient, GetParameterCommand } = require('@aws-sdk/client-ssm');

const app = express();
const PORT = process.env.PORT || 3000;

const register = new promClient.Registry();
promClient.collectDefaultMetrics({ register });

const httpRequestDuration = new promClient.Histogram({
  name: 'http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'route', 'status_code'],
  registers: [register]
});

const httpRequestTotal = new promClient.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status_code'],
  registers: [register]
});

app.use(express.json());
app.use(morgan('combined'));

let API_KEY = process.env.API_KEY;

const ssmClient = new SSMClient({ region: process.env.AWS_REGION || 'us-east-1' });

async function loadApiKey() {
  if (!API_KEY) {
    try {
      const command = new GetParameterCommand({
        Name: '/app/microservices-cicd/api-key',
        WithDecryption: true
      });
      const response = await ssmClient.send(command);
      API_KEY = response.Parameter.Value;
      console.log('API key loaded from Parameter Store');
    } catch (error) {
      console.error('Failed to load API key:', error);
    }
  }
}

loadApiKey();

const authMiddleware = (req, res, next) => {
  const apiKey = req.headers['x-api-key'];
  if (!apiKey || apiKey !== API_KEY) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  next();
};

const metricsMiddleware = (req, res, next) => {
  const start = Date.now();
  res.on('finish', () => {
    const duration = (Date.now() - start) / 1000;
    httpRequestDuration.labels(req.method, req.route?.path || req.path, res.statusCode).observe(duration);
    httpRequestTotal.labels(req.method, req.route?.path || req.path, res.statusCode).inc();
  });
  next();
};

app.use(metricsMiddleware);

const USER_SERVICE = process.env.USER_SERVICE_URL || 'http://user-service:8000';
const PRODUCT_SERVICE = process.env.PRODUCT_SERVICE_URL || 'http://product-service:8080';

app.get('/health', (req, res) => {
  res.json({ status: 'healthy', service: 'api-gateway' });
});

app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType);
  res.end(await register.metrics());
});

app.use('/users', authMiddleware, async (req, res) => {
  try {
    const response = await axios({
      method: req.method,
      url: `${USER_SERVICE}${req.path}`,
      data: req.body,
      params: req.query,
      headers: { 'Content-Type': 'application/json' }
    });
    res.status(response.status).json(response.data);
  } catch (error) {
    const status = error.response?.status || 500;
    res.status(status).json({ error: error.response?.data || 'User service error' });
  }
});

app.use('/products', authMiddleware, async (req, res) => {
  try {
    const response = await axios({
      method: req.method,
      url: `${PRODUCT_SERVICE}${req.path}`,
      data: req.body,
      params: req.query,
      headers: { 'Content-Type': 'application/json' }
    });
    res.status(response.status).json(response.data);
  } catch (error) {
    const status = error.response?.status || 500;
    res.status(status).json({ error: error.response?.data || 'Product service error' });
  }
});

app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

app.listen(PORT, () => {
  console.log(`API Gateway listening on port ${PORT}`);
});
