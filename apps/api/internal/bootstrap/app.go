package bootstrap

import (
	"net/http"

	"github.com/haseg/anifusion-canvas/apps/api/internal/config"
	"github.com/haseg/anifusion-canvas/apps/api/internal/http/handler"
	"github.com/haseg/anifusion-canvas/apps/api/internal/http/router"
	dbinfra "github.com/haseg/anifusion-canvas/apps/api/internal/infrastructure/db"
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
	e.HTTPErrorHandler = handler.JSONErrorHandler
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{cfg.FrontendOrigin},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
	}))

	studioStore := usecase.StudioStore(usecase.NewMemoryStudioStore())
	if cfg.StudioStore == "database" {
		store, err := dbinfra.NewStudioStore(cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		studioStore = store
	}
	studioService := usecase.NewStudioServiceWithStore(studioStore)
	studioHandler := handler.NewStudioHandler(studioService)
	healthHandler := handler.NewHealthHandler(dependency.NewChecker(cfg))
	router.Register(e, studioHandler, healthHandler)

	return &App{e: e, cfg: cfg}, nil
}

func (a *App) Start() error {
	return a.e.Start(":" + a.cfg.Port)
}
