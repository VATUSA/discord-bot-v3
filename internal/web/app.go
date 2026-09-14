package web

import (
	"context"
	"github.com/VATUSA/discord-bot-v3/internal/commands"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"regexp"
	"time"
)

var discordIdPattern = regexp.MustCompile(`^[0-9]{1,20}$`)

func App(cmds chan<- commands.Command, isReady func() bool) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://vatusa.net"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// API routes
	e.GET("/assignRoles/:id", assignRoles(cmds))
	e.POST("/assignRoles/:id", assignRoles(cmds))

	e.GET("/healthz", healthz(isReady))

	return e
}

// Run serves e on addr until ctx is cancelled, then shuts it down gracefully.
func Run(ctx context.Context, addr string, e *echo.Echo) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- e.Start(addr)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return e.Shutdown(shutdownCtx)
	}
}

func assignRoles(cmds chan<- commands.Command) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if !discordIdPattern.MatchString(id) {
			return c.JSON(http.StatusBadRequest, nil)
		}
		// Never block the request on the bot; reject if the backlog is full.
		select {
		case cmds <- commands.SyncMember{UserID: id}:
			return c.JSON(http.StatusOK, nil)
		default:
			log.Printf("Command queue full, dropping sync for member: %s", id)
			return c.JSON(http.StatusServiceUnavailable, nil)
		}
	}
}

func healthz(isReady func() bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !isReady() {
			return c.NoContent(http.StatusServiceUnavailable)
		}
		return c.NoContent(http.StatusOK)
	}
}
