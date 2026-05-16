import { createRootRoute, createRoute, createRouter, Outlet, redirect } from '@tanstack/react-router'
import { AppShell } from '@/components/layout/app-shell'
import { ExportRoute } from '@/routes/export'
import { GenerateRoute } from '@/routes/step1.generate'
import { InpaintRoute } from '@/routes/step2.inpaint'
import { EditRoute } from '@/routes/step3.edit'

function RootLayout() {
  return (
    <AppShell>
      <Outlet />
    </AppShell>
  )
}

const rootRoute = createRootRoute({
  component: RootLayout,
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({ to: '/step1' })
  },
})

const step1Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/step1',
  component: GenerateRoute,
})

const step2Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/step2',
  component: InpaintRoute,
})

const step3Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/step3',
  component: EditRoute,
})

const exportRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/export',
  component: ExportRoute,
})

const routeTree = rootRoute.addChildren([indexRoute, step1Route, step2Route, step3Route, exportRoute])

export const router = createRouter({
  routeTree,
})
