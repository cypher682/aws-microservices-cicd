FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install --omit=dev

FROM node:18-alpine

WORKDIR /app

RUN addgroup -g 1001 -S nodejs && adduser -S nodejs -u 1001

COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --chown=nodejs:nodejs . .

USER nodejs

EXPOSE 3000

CMD ["node", "src/index.js"]
