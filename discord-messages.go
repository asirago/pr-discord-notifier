package main

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func (app *application) sendOpenedPullRequestMessage(payload Payload) error {

	hexColor := "#46954A"
	color, err := strconv.ParseInt(hexColor[1:], 16, 64)
	if err != nil {
		return err
	}

	message := discordgo.MessageEmbed{
		Color: int(color),
		Description: addIssueURLToPullRequestBody(fmt.Sprintf(`
		[%s](%s)

		%s`, payload.PullRequest.Title, payload.PullRequest.HTMLURL, payload.PullRequest.Body), payload.Repo.HTMLURL),
		Author: &discordgo.MessageEmbedAuthor{
			URL:     payload.PullRequest.User.HTMLURL,
			IconURL: payload.PullRequest.User.AvatarURL,
			Name:    fmt.Sprintf("%s has opened a pull request", payload.PullRequest.User.Login),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.ibb.co/1mQMPC7/pr-1.png",
		},
	}

	_, err = app.dgo.ChannelMessageSendEmbed(app.cfg.ChannelID, &message)
	if err != nil {
		return err
	}

	return nil
}
