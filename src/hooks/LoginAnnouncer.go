package hooks

import (
	"assistbot/src"
	"log"
)

var LoginAnnouncer src.SessionHook = func(s src.Session, r src.SessionReady) {
	log.Println("Logged in as " + s.State.User.Username + " (" + s.State.User.ID + ")")
}
