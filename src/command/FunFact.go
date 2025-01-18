package command

import (
	"assistbot/global"
	"assistbot/src"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type funFactResult struct {
	Id         string `json:"id"`
	Text       string `json:"text"`
	Source     string `json:"source"`
	Source_url string `json:"source_url"`
	Language   string `json:"language"`
	Permalink  string `json:"permalink"`
}

const funFactHost = "uselessfacts.jsph.pl"
const funFactUrl = "https://" + funFactHost + "/api/v2/facts/random"

var FunFact = src.Command{
	Info: src.CmdInfo{
		Name:        "funfact",
		Description: "send random fun facts about.. anything",
	},
	Fn: func(opt src.CmdResFnArgs) error {
		req, err := global.NewHttpRequest(http.MethodGet, funFactUrl, nil)
		if err != nil {
			return err
		}
		res, err := global.HttpClient.Do(req)
		if err != nil {
			return errors.New(funFactHost + " dosent give any bytes to us")
		} else {
			defer res.Body.Close()
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.New("failed to get data as text from " + funFactHost)
		}
		data := &funFactResult{}
		err = json.Unmarshal(body, data)
		if err != nil {
			return errors.New("cant parse the result (as JSON) from " + funFactHost)
		}
		embed := &discordgo.MessageEmbed{
			Title:       "Did you know?",
			Description: data.Text,
			Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Powered by %s", funFactHost)},
			Color:       0x00ff00, // Green color
		}
		return opt.Result(&src.CmdResData{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
	},
}
