"use client";

import useSWR from "swr";
import { env } from "@/lib/config";
import { jsonFetcher } from "@/lib/fetcher";

type HealthResponse = {
  status: string;
  service: string;
  timestamp: string;
};

export function BackendStatus() {
  const endpoint = `${env.NEXT_PUBLIC_API_BASE_URL}/health`;

  const { data, error, isLoading } = useSWR<HealthResponse>(endpoint, jsonFetcher, {
    refreshInterval: 3000,
  });

  return (
    <div className="mt-6 rounded-xl border border-slate-200 bg-slate-50 p-3">
      <p className="text-xs font-semibold tracking-wide text-slate-500">BACKEND STATUS</p>
      {isLoading && <p className="mt-1 text-sm text-slate-700">確認中...</p>}
      {error && <p className="mt-1 text-sm text-rose-700">接続失敗: {error.message}</p>}
      {data && (
        <div className="mt-1 text-sm text-slate-700">
          <p>service: {data.service}</p>
          <p>status: {data.status}</p>
          <p>time: {new Date(data.timestamp).toLocaleTimeString("ja-JP")}</p>
        </div>
      )}
    </div>
  );
}
