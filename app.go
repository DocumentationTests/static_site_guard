package main

import (
	"github.com/g3ortega/hugo-auth/authentication"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/joho/godotenv"
	"github.com/markbates/goth/providers/github"
	"github.com/shareed2k/goth_fiber"
	"log"
	"os"
	"time"

	"github.com/markbates/goth"

	"github.com/gofiber/template/html"
)

func main() {
	godotenv.Load()
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), os.Getenv("GITHUB_CALLBACK")),
	)

	//// Initialize default config
	storage := sqlite3.New(sqlite3.Config{
		Database:        "./static_site_guard.sqlite3",
		Table:           "static_site_guard",
		Reset:           false,
		GCInterval:      10 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: 1 * time.Second,
	})

	// optional config
	config := session.Config{
		Storage:        storage,
		CookieSameSite: "Lax",
	}

	// create session handler
	store := session.New(config)

	goth_fiber.SessionStore = store

	app.Use(logger.New())

	app.Use(func(ctx *fiber.Ctx) error {
		return authentication.SessionHandler(ctx, store)
	})

	app.Static("/", "./content/public", fiber.Static{
		Browse: true,
	})

	app.Get("/login/:provider", goth_fiber.BeginAuthHandler)

	app.Get("/auth/callback/:provider", func(ctx *fiber.Ctx) error {
		return authentication.Callback(ctx, store)
	})

	app.Get("/login", func(ctx *fiber.Ctx) error {
		return ctx.Render("index", fiber.Map{})
	})

	app.Get("/not_authorized", func(ctx *fiber.Ctx) error {
		return ctx.Render("not_authorized", fiber.Map{})
	})

	app.Get("/logout", func(ctx *fiber.Ctx) error {
		return authentication.Logout(ctx, store)
	})

	if err := app.Listen(":8088"); err != nil {
		log.Fatal(err)
	}
}
