package hooks

import (
	"assistbot/global/env"
	"assistbot/src"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var TempWebpage src.LoadHook = func(s src.Session) {
	if !env.EnableTempWebserver {
		return
	}
	log.Println("-- Making TempWP at http://localhost:" + env.Port + "/ --")
	go func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "halo! for any info regarding assistbot, contact: github.com/vintheweirdass")
		})
		server := &http.Server{Addr: ":" + env.Port, Handler: handler}
		go func() {
			if err := server.ListenAndServe(); err != nil {
				log.Println("[TempWP]: Webpage stopped")
			}
		}()
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		<-stop

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			// TODO
		}
	}()
}
