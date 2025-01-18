package global

import (
	"embed"

	"github.com/vincent-petithory/dataurl"
)

//go:embed assets/logo.png
var LogoFile embed.FS
var LogoBytes, _ = LogoFile.ReadFile("logo.png")
var LogoDataUrl = dataurl.EncodeBytes(LogoBytes)
