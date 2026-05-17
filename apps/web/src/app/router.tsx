import { createRootRoute, createRoute, createRouter, Outlet, redirect } from '@tanstack/react-router'
import { AppShell } from '@/components/layout/app-shell'
import { GenerateRoute } from '@/routes/step1.generate'
import { InpaintRoute } from '@/routes/step2.inpaint'

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

const legacyStep3Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/step3',
  beforeLoad: () => {
    throw redirect({ to: '/step1' })
  },
})

const legacyStep3EditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/step3/edit',
  beforeLoad: () => {
    throw redirect({ to: '/step1' })
  },
})

const legacyExportRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/export',
  beforeLoad: () => {
    throw redirect({ to: '/step1' })
  },
})

const routeTree = rootRoute.addChildren([
  indexRoute,
  step1Route,
  step2Route,
  legacyStep3Route,
  legacyStep3EditRoute,
  legacyExportRoute,
])

export const router = createRouter({
  routeTree,
})
