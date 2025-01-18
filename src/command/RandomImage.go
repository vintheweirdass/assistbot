package command

import (
	"assistbot/src"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var RandomImage = src.Command{
	Info: src.CmdInfo{
		Name:        "random-image",
		Description: "random image from picsum.photos",
		Options: src.CmdInfoOpt{
			{
				Name:        "width",
				Description: "image width",
				Type:        src.CmdInfoOptTypeEnum.Integer,
				Required:    false,
			},
			{
				Name:        "height",
				Description: "image height",
				Type:        src.CmdInfoOptTypeEnum.Integer,
				Required:    false,
			},
			{
				Name:        "widthnheight",
				Description: "image width n height (may override the default `width` & `height` option)",
				Type:        src.CmdInfoOptTypeEnum.Integer,
				Required:    false,
			},
			{
				Name:        "blur",
				Description: "blur an image. for default, just set to 0",
				Type:        src.CmdInfoOptTypeEnum.Integer,
				Required:    false,
			},
			{
				Name:        "grayscale",
				Description: "make the image black & white",
				Type:        src.CmdInfoOptTypeEnum.Boolean,
				Required:    false,
			},
		},
	},
	Fn: func(opt src.CmdResFnArgs) error {
		width := "300"
		height := "300"

		var widthArg = opt.Args["width"]
		var heightArg = opt.Args["height"]
		var widthnheightArg = opt.Args["widthnheight"]
		var blurArg = opt.Args["blur"]
		var grayscaleArg = opt.Args["grayscale"]

		var params = "?random=2"

		if blurArg != nil {
			var blur = blurArg.IntValue()
			if blur == 0 {
				params += "&blur"
			} else {
				params += "&blur=" + fmt.Sprintf("%v", blur)
			}
		}
		if grayscaleArg != nil {
			if grayscaleArg.BoolValue() {
				params += "&grayscale"
			}
		}

		if widthArg != nil {
			width = fmt.Sprintf("%v", widthArg.IntValue())
		}
		if heightArg != nil {
			height = fmt.Sprintf("%v", heightArg.IntValue())
		}
		if widthnheightArg != nil {
			width = fmt.Sprintf("%v", widthnheightArg.IntValue())
			height = fmt.Sprintf("%v", widthnheightArg.IntValue())
		}
		log.Println("https://picsum.photos/" + width + "/" + height + params)
		embed := &discordgo.MessageEmbed{
			Title: "Here's the random image!",
			Image: &discordgo.MessageEmbedImage{
				URL: "https://picsum.photos/" + width + "/" + height + params,
			},
			Color: 0x00ff00, // Green color
		}
		return opt.Result(&src.CmdResData{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
	},
}
