package main

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

/*
   type Payload struct {
   	Action      string `json:"action"`
   	Number      int64  `json:"number"`
   	PullRequest struct {
   		URL     string `json:"url"`
   		ID      int64  `json:"id"`
   		HTMLURL string `json:"html_url"`
   		State   string `json:"state"`
   		Title   string `json:"title"`
   		User    struct {
   			Login string `json:"login"`
   			Title string `json:"title"`
   		} `json:"user"`
   		Body string `json:"body"`
   	} `json:"pull_request"`
   }
*/

func (app *application) sendOpenedPullRequestMessage(payload Payload) error {
	fmt.Println("Sending opened pull request message")

	hexColor := "#46954A"
	color, err := strconv.ParseInt(hexColor[1:], 16, 64)
	if err != nil {
		return err
	}

	message := discordgo.MessageEmbed{
		Color: int(color),
		Description: fmt.Sprintf(`
		[%s](%s)

		%s`, payload.PullRequest.Title, payload.PullRequest.HTMLURL, payload.PullRequest.Body),
		Author: &discordgo.MessageEmbedAuthor{
			URL:     payload.PullRequest.User.HTMLURL,
			IconURL: payload.PullRequest.User.AvatarURL,
			Name:    payload.PullRequest.User.Login,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/8/87/Octicons-git-pull-request.svg/1200px-Octicons-git-pull-request.svg.png",
		},
	}

	_, err = app.dgo.ChannelMessageSendEmbed(app.cfg.ChannelID, &message)
	if err != nil {
		return err
	}

	return nil
}
