import math
import os

import numpy as np
import onnxruntime as ort
from transformers import AutoTokenizer
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from pydantic import BaseModel


class Embedder:
    def __init__(self):
        model_path = os.getenv(
            "EMBEDDER_MODEL_PATH",
            "models/e5-multilingual-large/onnx/model_qint8_avx512_vnni.onnx"
        )

        tokenizer_path = os.getenv(
            "EMBEDDER_TOKENIZER_PATH",
            "models/e5-multilingual-large/onnx"
        )

        max_length = int(os.getenv("EMBEDDER_MAX_LEN", "128"))

        self.max_length = max_length
        self.tokenizer = AutoTokenizer.from_pretrained(tokenizer_path)

        self.session = ort.InferenceSession(
            model_path,
            providers=["CPUExecutionProvider"]
        )

        inputs = self.session.get_inputs()
        outputs = self.session.get_outputs()

        if len(inputs) < 2:
            raise RuntimeError("ONNX требует input_ids и attention_mask.")

        self.input_ids_name = inputs[0].name
        self.attn_mask_name = inputs[1].name
        self.output_name = outputs[0].name

    def embed(self, text: str) -> list[float]:
        text = text.strip()
        if not text:
            return []

        tokens = self.tokenizer(
            text,
            return_tensors="np",
            padding="max_length",
            truncation=True,
            max_length=self.max_length,
        )

        outputs = self.session.run(
            [self.output_name],
            {
                self.input_ids_name: tokens["input_ids"],
                self.attn_mask_name: tokens["attention_mask"],
            },
        )

        emb = outputs[0]  # np.ndarray

        # Ожидаемые варианты:
        # (1, seq_len, dim) -> токеновые эмбеддинги
        # (1, dim)          -> уже pooled эмбеддинг
        # (dim,)            -> совсем простой случай
        emb = np.array(emb)

        if emb.ndim == 3:
            # (batch, seq_len, dim)
            token_embs = emb[0]                 # (seq_len, dim)
            mask = tokens["attention_mask"][0]  # (seq_len,)
            mask = mask.astype(np.float32)

            # избегаем деления на 0
            if mask.sum() == 0:
                pooled = token_embs.mean(axis=0)
            else:
                mask = mask[:, None]            # (seq_len, 1)
                pooled = (token_embs * mask).sum(axis=0) / mask.sum()

            vec = pooled
        elif emb.ndim == 2:
            # (batch, dim)
            vec = emb[0]
        elif emb.ndim == 1:
            vec = emb
        else:
            raise RuntimeError(f"Неожиданная форма выхода ONNX: {emb.shape}")

        vec_list = vec.tolist()

        # защитимся от NaN/inf
        cleaned = []
        for x in vec_list:
            # тут x уже должен быть числом
            if not isinstance(x, (int, float)):
                raise TypeError(f"Элемент эмбеддинга не число, а {type(x)}: {x}")
            cleaned.append(0.0 if not math.isfinite(x) else float(x))

        return cleaned


embedder = Embedder()

app = FastAPI(title="RedCat Embedder (ONNX)")


class EmbedRequest(BaseModel):
    text: str


class EmbedResponse(BaseModel):
    vector: list[float]
    dim: int


@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest):
    text = req.text.strip()
    if not text:
        return {"vector": [], "dim": 0}

    vec = embedder.embed(text)
    return JSONResponse({
        "vector": vec,
        "dim": len(vec)
    })
