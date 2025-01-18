package command

import (
	"assistbot/global"
	"assistbot/src"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type garyResult struct {
	Url string `json:"url"`
}

const garyHost = "garybot.dev"
const garyUrl = "https://" + garyHost + "/api/gary"

var Gary = src.Command{
	Info: src.CmdInfo{
		Name:        "gary",
		Description: "Send random pics of Gary the cat",
	},
	Fn: func(opt src.CmdResFnArgs) error {
		req, err := global.NewHttpRequest(http.MethodGet, garyUrl, nil)
		if err != nil {
			return err
		}
		res, err := global.HttpClient.Do(req)
		if err != nil {
			return errors.New(garyHost + " dosent give any bytes to us")
		} else {
			defer res.Body.Close()
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.New("failed to get data as text from " + garyHost)
		}
		data := &garyResult{}
		err = json.Unmarshal(body, data)
		if err != nil {
			return errors.New("cant parse the result (as JSON) from " + garyHost)
		}
		if data.Url == "" {
			return errors.New(garyHost + " dosent give any direct link to the image")
		}
		embed := &discordgo.MessageEmbed{
			Title: "Here's your Gary picture!",
			Image: &discordgo.MessageEmbedImage{
				URL: data.Url,
			},
			Color: 0x00ff00, // Green color
		}
		return opt.Result(&src.CmdResData{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
	},
}
