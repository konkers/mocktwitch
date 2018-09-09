package mocktwitch

import (
	"bufio"
	"crypto/tls"
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
			fmt.Fprintf(conn, ":tmi.twitch.tv 001 justinfan123123 :Welcome, GLHF!\r\n")
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
	host := ":" + strconv.Itoa(listenPort)
	listenPort++

	cert, err := tls.LoadX509KeyPair(t.keys.certFilename, t.keys.keyFilename)
	if err != nil {
		return err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listener, err := tls.Listen("tcp", host, config)
	if err != nil {
		return err
	}

	t.IrcHost = host

	go t.serveIrc(listener)

	return nil
}
