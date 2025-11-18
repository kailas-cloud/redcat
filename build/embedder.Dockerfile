FROM python:3.11-slim

WORKDIR /app

ARG MODEL_DIR="models/e5-multilingual-large/onnx"
ARG MODEL_FILE="model_qint8_avx512_vnni.onnx"

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

COPY embedder ./embedder

RUN pip install --no-cache-dir -r ./embedder/requirements.txt


RUN mkdir -p /app/models

COPY ${MODEL_DIR}/${MODEL_FILE} /app/models/model.onnx

COPY ${MODEL_DIR}/tokenizer.json \
     ${MODEL_DIR}/tokenizer_config.json \
     ${MODEL_DIR}/sentencepiece.bpe.model \
     ${MODEL_DIR}/special_tokens_map.json \
     /app/models/

ENV PYTHONPATH=/app

EXPOSE 8000

CMD ["uvicorn", "embedder.app:app", "--host", "0.0.0.0", "--port", "8000"]
