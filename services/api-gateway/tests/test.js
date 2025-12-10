const request = require('supertest');
const express = require('express');

const app = express();
app.get('/health', (req, res) => {
  res.json({ status: 'healthy', service: 'api-gateway' });
});

describe('API Gateway', () => {
  describe('GET /health', () => {
    it('should return healthy status', async () => {
      const response = await request(app).get('/health');
      expect(response.status).toBe(200);
      expect(response.body.status).toBe('healthy');
      expect(response.body.service).toBe('api-gateway');
    });
  });
});
