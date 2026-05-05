const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8000";

type RequestOptions = {
  method?: string;
  body?: unknown;
};

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers: options.body ? { "Content-Type": "application/json" } : undefined,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });
  if (!res.ok) {
    const error = await res.text();
    throw new Error(`API error ${res.status}: ${error}`);
  }
  return res.json() as Promise<T>;
}

// --- Types ---

export interface Project {
  id: number;
  user_id: number;
  name: string;
  description: string | null;
  width: number;
  height: number;
  fps: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Frame {
  id: number;
  project_id: number;
  frame_index: number;
  is_keyframe: boolean;
  source_image_url: string | null;
  created_at: string;
  updated_at: string;
}

export interface Job {
  id: number;
  project_id: number;
  frame_id: number | null;
  type: string;
  status: string;
  result_url: string | null;
  error_message: string | null;
  created_at: string;
  updated_at: string;
}

// --- Health ---

export const health = () => request<{ status: string }>("/health");

// --- Projects ---

export const listProjects = () => request<Project[]>("/api/projects");
export const getProject = (id: number) => request<Project>(`/api/projects/${id}`);
export const createProject = (data: Partial<Project>) =>
  request<Project>("/api/projects", { method: "POST", body: data });

// --- Frames ---

export const listFrames = (projectId: number) =>
  request<Frame[]>(`/api/projects/${projectId}/frames`);

// --- Jobs ---

export const createInferenceJob = (projectId: number, image1Url: string, image2Url: string, prompt: string) =>
  request<Job>("/api/jobs/inference", {
    method: "POST",
    body: { project_id: projectId, image_1_url: image1Url, image_2_url: image2Url, prompt },
  });

export const createInpaintingJob = (projectId: number, frameId: number, maskUrl: string, prompt: string) =>
  request<Job>("/api/jobs/inpainting", {
    method: "POST",
    body: { project_id: projectId, frame_id: frameId, mask_url: maskUrl, prompt },
  });

export const createExportJob = (projectId: number) =>
  request<Job>("/api/jobs/export", {
    method: "POST",
    body: { project_id: projectId },
  });

export const getJob = (jobId: number) => request<Job>(`/api/jobs/${jobId}`);
