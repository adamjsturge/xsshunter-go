package main

import (
	"context"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/discord"
	"github.com/nikoksr/notify/service/msteams"
	"github.com/nikoksr/notify/service/slack"
	"github.com/nikoksr/notify/service/telegram"
)

func send_notification(title string, message string) {
	notify.Send(context.Background(), title, message)
}

func initialize_notify() {
	notify := notify.New()

	NOTIFY_KEY := get_env("NOTIFY_KEY")
	if NOTIFY_KEY == "" {
		return
	}

	switch get_env("NOTIFY_PROVIDER") {
	case "slack":
		slack_service := slack.New(NOTIFY_KEY)
		notify.UseServices(slack_service)
	case "discord":
		discord_service := discord.New()
		discord_service.AuthenticateWithBotToken(NOTIFY_KEY)
		notify.UseServices(discord_service)
	case "telegram":
		telegram_service, _ := telegram.New(NOTIFY_KEY)
		notify.UseServices(telegram_service)
	case "msteams":
		msteams_service := msteams.New()
		msteams_service.AddReceivers(NOTIFY_KEY)
		notify.UseServices(msteams_service)
	default:
	}

}
