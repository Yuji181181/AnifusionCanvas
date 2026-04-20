import { BackendStatus } from "@/components/backend-status";
import { CanvasEditor } from "@/components/canvas-editor";

export default function Home() {
  return (
    <main className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <section className="grid gap-6 lg:grid-cols-[320px_1fr]">
        <aside className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h1 className="text-2xl font-semibold text-slate-900">Anifusion Canvas</h1>
          <p className="mt-2 text-sm text-slate-600">
            AI中割り生成、フレーム単位のマスク修正、手動レタッチを統合するデモ環境です。
          </p>
          <div className="mt-4 space-y-2 text-sm text-slate-700">
            <p>1. キーフレームをアップロード</p>
            <p>2. 生成された各フレームを個別に修正</p>
            <p>3. 編集後に手動で動画を書き出す</p>
          </div>
          <p className="mt-4 rounded-lg bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-900">
            動画化は自動では行いません。フレーム編集が完了したあとに、書き出しボタンを押す運用にします。
          </p>
          <BackendStatus />
        </aside>
        <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:p-6">
          <CanvasEditor />
        </section>
      </section>
    </main>
  );
}
