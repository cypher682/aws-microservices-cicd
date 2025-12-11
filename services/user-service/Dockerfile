FROM python:3.11-slim

WORKDIR /app

RUN addgroup --system --gid 1001 appuser && adduser --system --uid 1001 --gid 1001 appuser

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY --chown=appuser:appuser . .

USER appuser

EXPOSE 8000

CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]
