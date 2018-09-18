package mocktwitch

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"

	"github.com/konkers/twitchapi/protocol"
)

type Twitch struct {
	IrcHost     string
	ApiUrlBase  string
	Errors      chan error
	ForceErrors bool

	IrcMeassageChan chan string

	// Keys is public so that downstream tests can use them.
	Keys    *Keys
	ircConn net.Conn

	ChannelStatus  protocol.Channel
	ChannelFollows protocol.ChannelFollows
}

var listenPort = 14823

func NewTwitch() (*Twitch, error) {
	t := &Twitch{
		Errors: make(chan error),
	}

	var err error
	t.Keys, err = generateCert("localhost")
	if err != nil {
		return nil, err
	}

	t.IrcMeassageChan = make(chan string, 100)

	err = t.newIrcServer()
	if err != nil {
		return nil, err
	}

	err = t.newAPIServer()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Twitch) Close() {
	os.Remove(t.Keys.CertFilename)
	os.Remove(t.Keys.KeyFilename)
}

func (t *Twitch) SendMessage(channel string, author string, message string) {
	encoded := fmt.Sprintf("@badges=subscriber/6,premium/1;color=#FF0000;display-name=%s;emotes=;id=2a31a9df-d6ff-4840-b211-a2547c7e656e;mod=0;room-id=11148817;subscriber=1;tmi-sent-ts=1490382457309;turbo=0;user-id=78424343;user-type= :%s!%s@%s.tmi.twitch.tv PRIVMSG #%s :%s\r\n",
		author, author, author, author, channel, message)

	fmt.Fprint(t.ircConn, encoded)
}

func (t *Twitch) SetChannelStatus(status *protocol.Channel) {
	t.ChannelStatus = *status
}

func (t *Twitch) getTLSListener(host string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(t.Keys.CertFilename, t.Keys.KeyFilename)
	if err != nil {
		return nil, err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return tls.Listen("tcp", host, config)
}
