FROM python:3.10-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    build-essential \
    python3-dev \
    libatlas-base-dev \
    libopenblas-dev \
    liblapack-dev \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir --upgrade pip \
    && pip install --no-cache-dir cython==0.29.36

COPY . .

RUN pip install --no-cache-dir -r requirements.txt --prefer-binary

EXPOSE 50054

CMD ["python", "grpc_server.py"]
