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

func (app *application) sendEditedPullRequestMessage(payload Payload) error {

	hexColor := "#FDE047"
	color, err := strconv.ParseInt(hexColor[1:], 16, 64)

	title := ""
	body := ""

	if payload.Changes.Title.From != "" {
		title = fmt.Sprintf("~~%s~~ \n", payload.Changes.Title.From)
	}

	if payload.Changes.Body.From != "" {
		body = fmt.Sprintf("~~%s~~\n\n", payload.Changes.Body.From)
	}

	message := discordgo.MessageEmbed{
		Color: int(color),
		Description: addIssueURLToPullRequestBody(fmt.Sprintf(`
		%s[%s](%s)

		%s%s`, title, payload.PullRequest.Title, payload.PullRequest.HTMLURL, body, payload.PullRequest.Body), payload.Repo.HTMLURL),
		Author: &discordgo.MessageEmbedAuthor{
			URL:     payload.PullRequest.HTMLURL,
			IconURL: payload.PullRequest.User.AvatarURL,
			Name:    fmt.Sprintf("%s has edited a pull request", payload.PullRequest.User.Login),
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

func (app *application) sendClosedPullRequestMessage(payload Payload) error {

	hexColor := "#8B5CF6"
	color, err := strconv.ParseInt(hexColor[1:], 16, 64)

	message := discordgo.MessageEmbed{
		Color: int(color),
		Description: addIssueURLToPullRequestBody(fmt.Sprintf(`
		[%s](%s)`, payload.PullRequest.Title, payload.PullRequest.HTMLURL), payload.Repo.HTMLURL),
		Author: &discordgo.MessageEmbedAuthor{
			URL:     payload.PullRequest.User.HTMLURL,
			IconURL: payload.PullRequest.User.AvatarURL,
			Name:    fmt.Sprintf("%s has closed a pull request", payload.PullRequest.User.Login),
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
