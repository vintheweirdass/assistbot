package command

import (
	"assistbot/src"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

type garyResult struct {
	Url string `json:"url"`
}

const host = "garybot.dev"
const GaryUrl = "https://" + host + "/api/gary"
const userAgent = "vintheweirdass-assistbot"

var garyClient = http.Client{
	Timeout: 6 * time.Second, // Timeout after 2 seconds
}
var Gary = src.Command{
	Info: src.CmdInfo{
		Name:        "gary",
		Description: "Send random pics of Gary the cat",
	},
	Fn: func(opt src.CmdResFnArgs) error {
		req, err := http.NewRequest(http.MethodGet, GaryUrl, nil)
		if err != nil {
			return err
		}
		req.Header.Set("user-Agent", userAgent)
		res, err := garyClient.Do(req)
		if err != nil {
			return errors.New(host + " dosent give any bytes to us")
		} else {
			defer res.Body.Close()
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.New("failed to get data as text from " + host)
		}
		data := &garyResult{}
		err = json.Unmarshal(body, data)
		if err != nil {
			return errors.New("cant parse the result (as JSON) from " + host)
		}
		if data.Url == "" {
			return errors.New(host + " dosent give any direct link to the image")
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
