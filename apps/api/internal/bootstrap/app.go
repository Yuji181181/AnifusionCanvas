package bootstrap

import (
	"net/http"

	"github.com/haseg/anifusion-canvas/apps/api/internal/config"
	"github.com/haseg/anifusion-canvas/apps/api/internal/http/handler"
	"github.com/haseg/anifusion-canvas/apps/api/internal/http/router"
	"github.com/haseg/anifusion-canvas/apps/api/internal/infrastructure/dependency"
	"github.com/haseg/anifusion-canvas/apps/api/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type App struct {
	e   *echo.Echo
	cfg config.Config
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{cfg.FrontendOrigin},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
	}))

	studioService := usecase.NewStudioService()
	studioHandler := handler.NewStudioHandler(studioService)
	healthHandler := handler.NewHealthHandler(dependency.NewChecker(cfg))
	router.Register(e, studioHandler, healthHandler)

	return &App{e: e, cfg: cfg}, nil
}

func (a *App) Start() error {
	return a.e.Start(":" + a.cfg.Port)
}
