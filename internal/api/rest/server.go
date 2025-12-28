package rest

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/stashapp/stash-box/internal/api/rest/handlers/edit"
	"github.com/stashapp/stash-box/internal/api/rest/handlers/performer"
	"github.com/stashapp/stash-box/internal/api/rest/handlers/scene"
	"github.com/stashapp/stash-box/internal/api/rest/handlers/studio"
	"github.com/stashapp/stash-box/internal/api/rest/handlers/tag"
	"github.com/stashapp/stash-box/internal/api/rest/handlers/system"
	"github.com/stashapp/stash-box/internal/api/rest/middleware"
	"github.com/stashapp/stash-box/internal/service"
)

// @title Stash-box REST API
// @version 1.0
// @description REST API for Stash-box metadata database
// @termsOfService http://swagger.io/terms/

// @contact.name Stash App
// @contact.url https://github.com/stashapp/stash-box

// @license.name AGPL-3.0
// @license.url https://www.gnu.org/licenses/agpl-3.0.en.html

// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name ApiKey
// @description API Key authentication

// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name stashbox
// @description Session cookie authentication

// SetupRESTRouter sets up the REST API router with all handlers and middleware
func SetupRESTRouter(fac service.Factory, version, hash, stamp string) chi.Router {
	r := chi.NewRouter()

	// Swagger UI
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/api/v1/docs/swagger.json"),
	))

	// System endpoints (no auth required for config/version)
	systemHandler := system.NewHandler(fac, version, hash, stamp)
	r.Mount("/system", systemHandler.Routes())

	// Scene endpoints
	sceneHandler := scene.NewHandler(fac)
	r.Mount("/scenes", sceneHandler.Routes())

	// Performer endpoints
	performerHandler := performer.NewHandler(fac)
	r.Mount("/performers", performerHandler.Routes())

	// Studio endpoints
	studioHandler := studio.NewHandler(fac)
	r.Mount("/studios", studioHandler.Routes())

	// Tag endpoints
	tagHandler := tag.NewHandler(fac)
	r.Mount("/tags", tagHandler.Routes())

	// Edit endpoints
	editHandler := edit.NewHandler(fac)
	r.Mount("/edits", editHandler.Routes())

	// TODO: Mount other handler packages as they are created

	// Users
	// r.Mount("/users", userHandler.Routes())
	// r.Mount("/auth", authHandler.Routes())

	// Drafts
	// r.Mount("/drafts", draftHandler.Routes())

	// Notifications
	// r.Mount("/notifications", notificationHandler.Routes())

	// Images
	// r.Mount("/images", imageHandler.Routes())

	// Sites
	// r.Mount("/sites", siteHandler.Routes())

	// Invites
	// r.Mount("/invites", inviteHandler.Routes())

	return r
}

// Ensure middleware is exported for use in handlers
var (
	RequireRole = middleware.RequireRole
	RequireAuth = middleware.RequireAuth
)
