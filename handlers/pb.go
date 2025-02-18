package handlers

import (
	"immodi/startup/auth"
	"immodi/startup/repo"
	"immodi/startup/routes"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func PocketBase() *pocketbase.PocketBase {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{os.Getenv("FRONTEND_URL"), "http://localhost:5173"},
			AllowHeaders: []string{
				echo.HeaderOrigin,
				echo.HeaderContentType,
				echo.HeaderAccept,
				echo.HeaderAuthorization,
				"Content-Disposition",
			},
			ExposeHeaders: []string{
				"Content-Disposition",
			},
			AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
			AllowCredentials: false,
		}))

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/hello",
			Handler: func(c echo.Context) error {
				return c.String(200, "Hello world!")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/generate",
			Handler: func(c echo.Context) error {
				return routes.Generate(c, app)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				// apis.RequireAdminOrRecordAuth("users"),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/auth",
			Handler: func(c echo.Context) error {
				return auth.AuthenticateUserWithUsername(c, app)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/admin-auth",
			Handler: func(c echo.Context) error {
				return auth.AuthenticateAdmin(c, app)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/templates",
			Handler: func(c echo.Context) error {
				return repo.GetUserTemplates(c, app)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				// apis.RequireRecordAuth("users"),
			},
		})

		return nil
	})

	app.OnRecordAfterCreateRequest("users").Add(func(e *core.RecordCreateEvent) error {
		userTemplates := []string{
			"waxxopaxrgdpkki",
			"8gnqdsso46yp6pm",
			"mqcpw4e0qdb0tg6",
		}

		e.Record.Set("user_templates", userTemplates)
		e.Record.Set("user_files", []string{})
		e.Record.Set("tokens", 50)
		e.Record.Set("current_plan", "kemt0gtyrxjahfh")

		err := app.Dao().Save(e.Record)
		if err != nil {
			return err
		}

		return nil
	})

	app.OnRecordAfterAuthWithOAuth2Request().Add(func(e *core.RecordAuthWithOAuth2Event) error {
		if e.IsNewRecord {
			userTemplates := []string{
				"waxxopaxrgdpkki",
				"8gnqdsso46yp6pm",
				"mqcpw4e0qdb0tg6",
			}
			e.Record.Set("user_templates", userTemplates)
			e.Record.Set("user_files", []string{})
			e.Record.Set("tokens", 50)
			e.Record.Set("current_plan", "kemt0gtyrxjahfh")

			err := app.Dao().Save(e.Record)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	return app
}
