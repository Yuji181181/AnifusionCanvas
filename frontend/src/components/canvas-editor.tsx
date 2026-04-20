"use client";

import { useMemo, useState } from "react";
import { Layer, Line, Rect, Stage } from "react-konva";
import { env } from "@/lib/config";
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
  const [jobId, setJobId] = useState("");
  const [exportMessage, setExportMessage] = useState("未書き出し");

  const brushLabel = useMemo(() => `${brushSize}px`, [brushSize]);

  const handleExport = async () => {
    if (!jobId.trim()) {
      setExportMessage("Job ID を入力してください");
      return;
    }

    setExportMessage("動画を書き出し中...");
    const response = await fetch(`${env.NEXT_PUBLIC_API_BASE_URL}/v1/jobs/${jobId}/export`, {
      method: "POST",
    });

    if (!response.ok) {
      const detail = await response.text();
      setExportMessage(`書き出し失敗: ${response.status} ${detail}`);
      return;
    }

    const data = (await response.json()) as { video_url: string };
    setExportMessage(`書き出し完了: ${data.video_url}`);
  };

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
          <div className="rounded-lg border border-dashed border-slate-300 bg-white p-3 text-xs leading-5 text-slate-600">
            ここではフレーム単位でマスクと修正内容を作ります。動画への書き出しは、細かい編集が終わってから手動で行います。
          </div>
          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700" htmlFor="job-id">
              書き出し対象 Job ID
            </label>
            <input
              id="job-id"
              type="text"
              value={jobId}
              onChange={(event) => setJobId(event.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
              placeholder="例: 123e4567-e89b-12d3-a456-426614174000"
            />
          </div>
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
            <button
              type="button"
              onClick={handleExport}
              className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-800 hover:bg-slate-100"
            >
              動画を書き出す
            </button>
          </div>
          <p className="text-xs leading-5 text-slate-500">{exportMessage}</p>
        </div>
      </div>
    </div>
  );
}
