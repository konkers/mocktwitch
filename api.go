package mocktwitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/konkers/twitchapi/protocol"

	"github.com/gorilla/mux"
)

func (t *Twitch) apiError(w http.ResponseWriter, req *http.Request, format string, args ...interface{}) {
	errStr := fmt.Sprintf(format, args...)
	http.Error(w, errStr, 500)
}

func (t *Twitch) apiHandleChannel(w http.ResponseWriter, req *http.Request) {
	if t.ForceErrors {
		t.apiError(w, req, "Forced Error")
		return
	}
	if req.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")

		b, err := json.Marshal(&t.ChannelStatus)
		if err != nil {
			t.apiError(w, req, "can't marshal connection: %v.", err)
			return
		}
		w.Write(b)
	}
}

func (t *Twitch) apiHandlePutChannels(w http.ResponseWriter, req *http.Request) {
	if t.ForceErrors {
		t.apiError(w, req, "Forced Error")
		return
	}
	vars := mux.Vars(req)
	channelName, ok := vars["channel"]
	if !ok {
		t.apiError(w, req, "No channel name var from mux!")
		return
	}

	if channelName != "test" {
		t.apiError(w, req, "Only test channel supported.")
		return
	}

	var update protocol.Update
	err := json.NewDecoder(req.Body).Decode(&update)
	if err != nil {
		t.apiError(w, req, "Can't decode update.")
		return
	}

	if update.Channel == nil {
		t.apiError(w, req, "No Channel data in update.")
		return
	}

	if update.Channel.Status != nil {
		t.ChannelStatus.Status = *update.Channel.Status
	}

	if update.Channel.Game != nil {
		t.ChannelStatus.Game = *update.Channel.Game
	}

	// Nothing to do with Delay or ChannelFeedEnabled right now.

	fmt.Fprintf(w, "OK")
}

func (t *Twitch) newAPIServer() error {
	host := ":" + strconv.Itoa(listenPort)
	listenPort++

	t.ApiUrlBase = "https://localhost" + host
	r := mux.NewRouter()
	r.HandleFunc("/channel", t.apiHandleChannel)
	r.HandleFunc("/channels/{channel}", t.apiHandlePutChannels).Methods("PUT")
	go http.ListenAndServeTLS(host, t.keys.certFilename, t.keys.keyFilename, r)
	return nil
}
