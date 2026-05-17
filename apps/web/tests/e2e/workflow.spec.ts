import { test, expect } from '@playwright/test'

test.describe('AnifusionCanvas E2E', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/projects/*/frames', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ frames: [] }),
      })
    })
  })

  test('app loads and shows main workflow', async ({ page }) => {
    await page.goto('/step1')

    await expect(page.locator('h1').first()).toBeVisible()
    await expect(page.locator('.step-nav')).toBeVisible()
    await expect(page.locator('.timeline')).toBeVisible()
  })

  test('step navigation works', async ({ page }) => {
    await page.goto('/')

    const stepLinks = page.locator('.step-link')
    await expect(stepLinks).toHaveCount(4)

    await stepLinks.nth(1).click()
    await expect(page).toHaveURL(/step2/)

    await stepLinks.nth(2).click()
    await expect(page).toHaveURL(/step3/)

    await stepLinks.nth(3).click()
    await expect(page).toHaveURL(/export/)
  })

  test('generation form validates input', async ({ page }) => {
    await page.goto('/step1')

    await expect(page.locator('h1')).toContainText('中割りを生成')

    const submitButton = page.getByRole('button', { name: /生成を開始/ })
    await expect(submitButton).toBeVisible()

    await page.locator('textarea').first().fill('')
    await submitButton.click()

    await expect(page.locator('.field-error').first()).toBeVisible()
  })

  test('timeline displays frame kind tags after generating demo frames', async ({ page }) => {
    await page.goto('/step1')

    await page.route('**/inference/generate', async (route) => {
      await route.fulfill({
        status: 202,
        contentType: 'application/json',
        body: JSON.stringify({
          job: {
            id: 'job-e2e-1',
            type: 'generation',
            status: 'queued',
            progress: 0,
            message: 'accepted',
            version: 1,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        }),
      })
    })

    await page.route('**/jobs/job-e2e-1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          job: {
            id: 'job-e2e-1',
            type: 'generation',
            status: 'succeeded',
            progress: 100,
            message: 'done',
            version: 2,
            result: {
              frames: [
                { id: 'f-0', projectId: 'demo-project', index: 0, imageUrl: 'data:image/png;base64,A', thumbnailUrl: 'data:image/png;base64,A', kind: 'key', updatedAt: new Date().toISOString() },
                { id: 'f-1', projectId: 'demo-project', index: 1, imageUrl: 'data:image/png;base64,B', thumbnailUrl: 'data:image/png;base64,B', kind: 'generated', updatedAt: new Date().toISOString() },
                { id: 'f-2', projectId: 'demo-project', index: 2, imageUrl: 'data:image/png;base64,C', thumbnailUrl: 'data:image/png;base64,C', kind: 'inpainted', updatedAt: new Date().toISOString() },
                { id: 'f-3', projectId: 'demo-project', index: 3, imageUrl: 'data:image/png;base64,D', thumbnailUrl: 'data:image/png;base64,D', kind: 'edited', updatedAt: new Date().toISOString() },
              ],
            },
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        }),
      })
    })

    const submitButton = page.getByRole('button', { name: /生成を開始/ })
    await submitButton.click()

    const frameThumbs = page.locator('.frame-thumb')
    await expect(frameThumbs).toHaveCount(4)

    await expect(page.locator('.frame-kind-generated')).toBeVisible()
    await expect(page.locator('.frame-kind-inpainted')).toBeVisible()
    await expect(page.locator('.frame-kind-edited')).toBeVisible()
  })

  test('editor panel renders canvas and tools', async ({ page }) => {
    await page.goto('/step3')

    await expect(page.locator('.toolbar').first()).toBeVisible()
    await expect(page.locator('canvas').first()).toBeVisible()
    await expect(page.locator('.icon-button').first()).toBeVisible()
  })

  test('inpainting target frame can be switched in step 2', async ({ page }) => {
    await page.goto('/step1')

    await page.route('**/inference/generate', async (route) => {
      await route.fulfill({
        status: 202,
        contentType: 'application/json',
        body: JSON.stringify({
          job: {
            id: 'job-e2e-picker',
            type: 'generation',
            status: 'queued',
            progress: 0,
            message: 'accepted',
            version: 1,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        }),
      })
    })

    await page.route('**/jobs/job-e2e-picker', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          job: {
            id: 'job-e2e-picker',
            type: 'generation',
            status: 'succeeded',
            progress: 100,
            message: 'done',
            version: 2,
            result: {
              frames: [
                { id: 'pick-0', projectId: 'demo-project', index: 0, imageUrl: 'data:image/png;base64,A', thumbnailUrl: 'data:image/png;base64,A', kind: 'key', updatedAt: new Date().toISOString() },
                { id: 'pick-1', projectId: 'demo-project', index: 1, imageUrl: 'data:image/png;base64,B', thumbnailUrl: 'data:image/png;base64,B', kind: 'generated', updatedAt: new Date().toISOString() },
              ],
            },
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        }),
      })
    })

    await page.getByRole('button', { name: /生成を開始/ }).click()
    await expect(page.locator('.frame-thumb')).toHaveCount(2)

    await page.locator('.step-link').nth(1).click()

    const options = page.locator('.target-frame-option')
    await expect(options).toHaveCount(2)
    await expect(options.first()).toHaveAttribute('aria-pressed', 'true')

    await options.nth(1).click()
    await expect(options.nth(1)).toHaveAttribute('aria-pressed', 'true')
  })
})
