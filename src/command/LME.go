package command

// lenovo model explorer. not affiliated in any way to Lenovo

import (
	"assistbot/global"
	"assistbot/src"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const lmeBscoHost = "download.lenovo.com"
const lmeSpecHost = "pcsupport.lenovo.com"
const lmeBscoUrl = "https://" + lmeBscoHost + "/bsco/public/allModels.json"
const lmeSpecUrl = "https://" + lmeSpecHost + "/us/en/api/v4/mse/getproducts?productId="
const lmeSpecHref = "https://" + lmeSpecHost + "/us/en/products/"

func LMEGetSpec(model string) (*global.LMESpecResult, error) {
	spec := &global.LMESpecResult{}
	// req, err := global.NewHttpRequest(http.MethodGet, lmeSpecUrl+model, nil)
	// if err != nil {
	// 	return spec, err
	// }
	res, err := global.HttpClient.Get(lmeSpecUrl + model)
	if err != nil {
		log.Print(err.Error(), lmeSpecUrl+model)
		return spec, errors.New(lmeSpecHost + " dosent give any bytes to us")
	} else {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return spec, errors.New("failed to get data as text from " + lmeSpecHost)
	}
	data := &[]*global.LMESpecResult{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return spec, errors.New("product `" + model + "` dosent found")
	}
	dataSlice := *data
	spec = dataSlice[0]
	return spec, nil
}

func LMERefreshBscoDB() error {
	log.Println("-- Loading LME BscoDB --")
	req, err := global.NewHttpRequest(http.MethodGet, lmeBscoUrl, nil)
	if err != nil {
		return err
	}
	res, err := global.HttpClient.Do(req)
	if err != nil {
		return errors.New(lmeBscoHost + " dosent give any bytes to us")
	} else {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.New("failed to get data as text from " + lmeBscoHost)
	}
	data := &[]*global.LMEBscoResult{}
	dataRes := data
	err = json.Unmarshal(body, data)
	if err != nil {
		return errors.New("cant parse the result (as JSON) from " + lmeBscoHost)
	}
	regex, err := regexp.Compile(`.+\([A-Z0-9]{4}(\,[A-Z0-9]{4})*\)`)
	if err != nil {
		return errors.New("internal error: failed to compile regex to match the model")
	}
	for _, e := range *data {
		match := regex.Match([]byte(e.Name))
		if !match {
			continue
		}
		dataTemp := append(*dataRes, e)
		dataRes = &dataTemp
	}
	global.BscoDB = *dataRes
	return nil
}

var LMEHook src.LoadHook = func(s src.Session) {
	var e = LMERefreshBscoDB()
	if e != nil {
		log.Println(e.Error())
	}
}

// masih ada bug di (LMEGetSpec)
var LME = src.Command{
	Info: src.CmdInfo{
		Name:        "lme",
		Description: "Lenovo Model Explorer. NOT AFFILIATED IN ANY WAY WITH LENOVO",
		Options: src.CmdInfoOpt{
			{
				Name:        "name",
				Type:        src.CmdInfoOptTypeEnum.String,
				Description: "model name. may override `query` argument",
				Required:    false,
			},
			{
				Name:        "query",
				Type:        src.CmdInfoOptTypeEnum.String,
				Description: "find models",
				Required:    false,
			},
		},
	},
	Fn: func(opt src.CmdResFnArgs) error {
		var nameArg = opt.Args["name"]
		if nameArg != nil {
			var name = nameArg.StringValue()
			idx := slices.IndexFunc(global.BscoDB, func(c *global.LMEBscoResult) bool {
				return strings.Contains(strings.ToLower(c.Name), strings.ToLower(name))
			})
			if idx < 0 {
				return errors.New("product `" + name + "` dosent found on " + lmeBscoHost)
			}
			var bsco = global.BscoDB[idx]
			var model = ""
			var takeModelStart = false
			for _, char := range bsco.Name {
				var s = string(char)
				if s == "(" {
					takeModelStart = true
					continue
				}
				if takeModelStart && (s == ")" || s == ",") {
					takeModelStart = false
					break
				}
				if takeModelStart {
					model = model + s
				}
			}
			if model == "" {
				return errors.New("product `" + name + "` dosent found on " + lmeBscoHost)
			}
			var spec, err = LMEGetSpec(model)
			if err != nil {
				return err
			}
			var genUrl = lmeSpecHref + strings.ToLower(spec.Id)
			embed := &discordgo.MessageEmbed{
				Type:  discordgo.EmbedTypeArticle,
				Title: spec.Name,
				Image: &discordgo.MessageEmbedImage{
					URL: spec.Image,
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Note: AssistBot is not affiliated in any way to Lenovo",
				},
				Color: 0x00ff00, // Green color
			}
			return opt.Result(&src.CmdResData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label: "Original website",
								Style: discordgo.LinkButton,
								URL:   genUrl,
							},
							discordgo.Button{
								Label: "Drivers & Software",
								Style: discordgo.LinkButton,
								URL:   genUrl + "/downloads",
							},
							discordgo.Button{
								Label: "Parts",
								Style: discordgo.LinkButton,
								URL:   genUrl + "/parts",
							},
							discordgo.Button{
								Label: "Userguide",
								Style: discordgo.LinkButton,
								URL:   genUrl + "/document-userguide",
							},
							discordgo.Button{
								Label: "How To's",
								Style: discordgo.LinkButton,
								URL:   genUrl + "/documentation",
							},
						},
					},
				},
			})
		}
		var queryArg = opt.Args["query"]
		const maxModelsShown = 8
		dest := make([]*global.LMEBscoResult, len(global.BscoDB))
		if queryArg != nil {
			var q = strings.ToLower(queryArg.StringValue())
			var idx = 0
			for _, v := range global.BscoDB {
				if !strings.Contains(strings.ToLower(v.Name), q) {
					continue
				}
				dest[idx] = v
				idx++
			}
		} else {
			perm := rand.Perm(len(global.BscoDB))
			for i, v := range perm {
				dest[v] = global.BscoDB[i]
			}
		}
		var result = ""
		if len(dest) < 1 {
			return errors.New("products dosent found")
		}
		for i, v := range dest {
			if i > maxModelsShown {
				break
			}
			var a = &v.Name
			result += string(*a) + "\n"
		}
		return opt.Result(&src.CmdResData{
			Content: "## List of models (max " + fmt.Sprintf("%d", maxModelsShown) + ") \n" + result,
		})
	},
}
