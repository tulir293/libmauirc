// libmauirc - An IRC connection library for mauIRCd
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	msg "github.com/sorcix/irc"
	"github.com/sorcix/irc/ctcp"
	irc "maunium.net/go/libmauirc"
	flag "maunium.net/go/mauflag"
)

var ip = flag.Make().ShortKey("a").LongKey("address").Usage("The address to connect to.").String()
var port = flag.Make().ShortKey("p").LongKey("port").Usage("The port to connect to.").Uint16()
var tls = flag.Make().ShortKey("s").LongKey("ssl").LongKey("tls").Usage("Whether or not to enable TLS.").Bool()
var wantHelp, _ = flag.MakeHelpFlag()

func main() {
	err := flag.Parse()
	flag.SetHelpTitles("lmitest - A simple program to test libmauirc.", "lmitest [-h] [-s] [-a IP ADDRESS] [-p PORT]")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.PrintHelp()
		os.Exit(1)
	} else if *wantHelp {
		flag.PrintHelp()
		os.Exit(0)
	}

	c := irc.Create("lmitest", "lmitest", irc.IPv4Address{IP: *ip, Port: *port})
	c.SetRealName("libmauirc tester")
	c.SetDebugWriter(os.Stdout)
	c.SetUseTLS(*tls)

	err = c.Connect()
	if err != nil {
		panic(err)
	}

	go c.Loop()

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	usr1 := make(chan os.Signal, 1)
	signal.Notify(usr1, syscall.SIGUSR1)
	usr2 := make(chan os.Signal, 1)
	signal.Notify(usr2, syscall.SIGUSR2)
	go func() {
		for {
			select {
			case <-usr1:
				c.Debugln("Disconnecting...")
				c.Disconnect()
			case <-term:
				c.Debugln("\nQuitting...")
				c.Quit()
				break
			case <-usr2:
				c.Debugln("Quitting...")
				c.Quit()
				break
			}
		}
	}()

	go func() {
		err := <-c.Errors()
		fmt.Fprintln(os.Stderr, "[Error]", err)
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		send := msg.ParseMessage(text)
		if strings.HasPrefix(send.Command, "CTCP_") {
			send.Trailing = ctcp.Encode(send.Command[len("CTCP_"):], send.Trailing)
			send.Command = msg.PRIVMSG
		}
		c.Send(send)
	}
}
