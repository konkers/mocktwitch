package mocktwitch

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

func (t *Twitch) serveIrc(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		t.Errors <- err
		return
	}

	t.ircConn = conn

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	defer listener.Close()
	defer conn.Close()
	for {
		message, err := tp.ReadLine()
		if err != nil {
			t.Errors <- err
			return
		}
		message = strings.Replace(message, "\r\n", "", 1)
		if strings.HasPrefix(message, "NICK") {
			if !t.SquelchIrc {
				fmt.Fprintf(conn, ":tmi.twitch.tv 001 justinfan123123 :Welcome, GLHF!\r\n")
			}
		} else {
			t.onIRCMessage(message)
		}
	}
}

func (t *Twitch) onIRCMessage(message string) {
	select {
	case t.IrcMeassageChan <- message:
	default:
		log.Printf("dropped message %s", message)
	}
}

func (t *Twitch) newIrcServer() error {
	host := "127.0.0.1:" + strconv.Itoa(listenPort)
	listenPort++
	t.IrcHost = host

	listener, err := t.getTLSListener(host)
	if err != nil {
		return err
	}

	go t.serveIrc(listener)

	return nil
}
