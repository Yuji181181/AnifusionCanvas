import { z } from 'zod'

export const generationFormSchema = z
  .object({
    projectId: z.string().min(1, 'プロジェクトIDは必須です'),
    prompt: z.string().min(1, 'プロンプトは必須です').max(500, 'プロンプトは500文字以内で入力してください'),
    negativePrompt: z.string().max(500).optional(),
    frameCount: z.number().int().min(2, '最低2フレーム必要です').max(12, '最大12フレームまでです'),
    startImageDataUrl: z.string().min(1, '開始フレームを選択してください'),
    endImageDataUrl: z.string().min(1, '終了フレームを選択してください'),
  })
  .refine((data) => data.startImageDataUrl !== data.endImageDataUrl, {
    message: '開始フレームと終了フレームは異なる画像を選択してください',
    path: ['endImageDataUrl'],
  })

export type GenerationFormValues = z.infer<typeof generationFormSchema>

export const inpaintingFormSchema = z.object({
  projectId: z.string().min(1, 'プロジェクトIDは必須です'),
  frameId: z.string().min(1, '編集するフレームを選択してください'),
  prompt: z.string().min(1, '修正内容を入力してください').max(500, '500文字以内で入力してください'),
  maskDataUrl: z.string().min(1, 'マスクが描画されていません'),
  strength: z.number().min(0.1, '0.1以上で指定してください').max(1, '1以下で指定してください'),
})

export type InpaintingFormValues = z.infer<typeof inpaintingFormSchema>

export const exportFormSchema = z.object({
  projectId: z.string().min(1, 'プロジェクトIDは必須です'),
  fps: z.number().int().min(1, '1以上で指定してください').max(60, '60以下で指定してください'),
})

export type ExportFormValues = z.infer<typeof exportFormSchema>
