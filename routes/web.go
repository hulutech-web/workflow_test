package routes

import (
	"goravel/packages/goravel-workflow/middleware"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/support"

	"goravel/app/facades"
	"goravel/app/http/controllers"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("welcome.tmpl", map[string]any{
			"version": support.Version,
		})
	})

	facades.Route().Static("public", "./public")
	facades.Route().Middleware(middleware.NewJwt()).Prefix("/api").Group(func(router route.Router) {
		userController := controllers.NewUserController()
		router.Resource("user", userController)
	})

}
