package main

import (
	"context"
	"github.com/VATUSA/discord-bot-v3/internal/bot"
	"github.com/VATUSA/discord-bot-v3/internal/commands"
	"github.com/VATUSA/discord-bot-v3/internal/config"
	"github.com/VATUSA/discord-bot-v3/internal/web"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const commandBufferSize = 100

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmds := make(chan commands.Command, commandBufferSize)

	webErr := make(chan error, 1)
	go func() {
		// If the web server dies, take the bot down with it so the pod gets restarted.
		webErr <- web.Run(ctx, config.HTTPAddr, web.App(cmds, bot.IsReady))
		stop()
	}()

	botErr := bot.Run(ctx, cmds)
	stop()

	failed := false
	if botErr != nil {
		log.Printf("Bot error: %v", botErr)
		failed = true
	}
	if err := <-webErr; err != nil {
		log.Printf("Web server error: %v", err)
		failed = true
	}
	if failed {
		os.Exit(1)
	}
}
