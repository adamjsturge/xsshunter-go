package main

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/types"
)

func send_notification(message string, screenshot_url string) {
	notify_urls := get_env("NOTIFY")
	if notify_urls == "" {
		return
	}

	message_with_screenshot := message + " " + screenshot_url

	urls := strings.Split(notify_urls, ",")

	sender, err := shoutrrr.CreateSender(urls...)
	params := &types.Params{}
	sender.Send(message_with_screenshot, params)
	if err != nil {
		fmt.Println("Error sending notification:", err)
	}
}
