"use client";

import { useMemo, useState } from "react";
import { Layer, Line, Rect, Stage } from "react-konva";
import { useEditorStore } from "@/store/editor-store";

type Stroke = {
  id: string;
  points: number[];
  strokeWidth: number;
};

const STAGE_WIDTH = 960;
const STAGE_HEIGHT = 540;

export function CanvasEditor() {
  const { brushSize, prompt, setBrushSize, setPrompt } = useEditorStore();
  const [strokes, setStrokes] = useState<Stroke[]>([]);
  const [isDrawing, setIsDrawing] = useState(false);

  const brushLabel = useMemo(() => `${brushSize}px`, [brushSize]);

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-[1fr_240px]">
        <div className="overflow-hidden rounded-xl border border-slate-300 bg-white">
          <Stage
            width={STAGE_WIDTH}
            height={STAGE_HEIGHT}
            onMouseDown={(event) => {
              setIsDrawing(true);
              const pointer = event.target.getStage()?.getPointerPosition();
              if (!pointer) {
                return;
              }

              setStrokes((prev) => [
                ...prev,
                {
                  id: crypto.randomUUID(),
                  points: [pointer.x, pointer.y],
                  strokeWidth: brushSize,
                },
              ]);
            }}
            onMouseMove={(event) => {
              if (!isDrawing) {
                return;
              }

              const pointer = event.target.getStage()?.getPointerPosition();
              if (!pointer) {
                return;
              }

              setStrokes((prev) => {
                const copy = [...prev];
                const last = copy[copy.length - 1];
                if (!last) {
                  return prev;
                }
                last.points = [...last.points, pointer.x, pointer.y];
                return copy;
              });
            }}
            onMouseUp={() => setIsDrawing(false)}
            onMouseLeave={() => setIsDrawing(false)}
          >
            <Layer>
              <Rect x={0} y={0} width={STAGE_WIDTH} height={STAGE_HEIGHT} fill="#f8fafc" />
              {strokes.map((stroke) => (
                <Line
                  key={stroke.id}
                  points={stroke.points}
                  stroke="#0f172a"
                  strokeWidth={stroke.strokeWidth}
                  tension={0.5}
                  lineCap="round"
                  lineJoin="round"
                  globalCompositeOperation="source-over"
                />
              ))}
            </Layer>
          </Stage>
        </div>
        <div className="space-y-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700" htmlFor="brush">
              ブラシサイズ ({brushLabel})
            </label>
            <input
              id="brush"
              type="range"
              min={4}
              max={72}
              value={brushSize}
              onChange={(event) => setBrushSize(Number(event.target.value))}
              className="w-full"
            />
          </div>
          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700" htmlFor="prompt">
              Inpaintingプロンプト
            </label>
            <input
              id="prompt"
              type="text"
              value={prompt}
              onChange={(event) => setPrompt(event.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
              placeholder="例: 握りこぶし"
            />
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setStrokes([])}
              className="rounded-lg bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800"
            >
              マスクをクリア
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
