FROM python:3.14-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

RUN groupadd --system --gid 10001 agentshield \
    && useradd \
       --system \
       --uid 10001 \
       --gid 10001 \
       --create-home \
       agentshield

COPY requirements.txt .

RUN pip install \
    --no-cache-dir \
    --upgrade pip \
    && pip install \
       --no-cache-dir \
       -r requirements.txt

COPY app ./app
COPY models ./models

RUN chown -R 10001:10001 /app

USER 10001:10001

EXPOSE 8085

CMD ["uvicorn", "app.api:app", "--host", "0.0.0.0", "--port", "8085"]