import { BackendStatus } from "@/components/backend-status";
import { CanvasEditor } from "@/components/canvas-editor";

export default function Home() {
  return (
    <main className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <section className="grid gap-6 lg:grid-cols-[320px_1fr]">
        <aside className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h1 className="text-2xl font-semibold text-slate-900">Anifusion Canvas</h1>
          <p className="mt-2 text-sm text-slate-600">
            AI中割り生成、マスク修正、手動レタッチを統合するデモ環境の初期セットアップです。
          </p>
          <div className="mt-4 space-y-2 text-sm text-slate-700">
            <p>1. キーフレームをアップロード</p>
            <p>2. 推論ジョブを実行</p>
            <p>3. マスクとプロンプトで再生成</p>
          </div>
          <BackendStatus />
        </aside>
        <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:p-6">
          <CanvasEditor />
        </section>
      </section>
    </main>
  );
}
