import math
import os

import onnxruntime as ort
from transformers import AutoTokenizer
from fastapi import FastAPI
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

        # имена входов/выходов
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

        vec = outputs[0][0]  # numpy array

        # Convert to native Python list - handles numpy types properly
        vec_list = vec.tolist()

        # Clean non-finite values
        cleaned = [0.0 if not math.isfinite(x) else x for x in vec_list]

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
    return {"vector": vec, "dim": len(vec)}
