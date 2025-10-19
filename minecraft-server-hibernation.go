package main

import (
	"flag"
	"fmt"
	"net"

	"msh/lib/config"
	"msh/lib/conn"
	"msh/lib/errco"
	"msh/lib/input"
	"msh/lib/progmgr"
	"msh/lib/servctrl"
)

var (
	logPrefix = flag.String("prefix", "Mineplus", "prefix for wrapper logs")
	quiet     = flag.Bool("quiet", true, "suppress non-error wrapper logs")
)

func main() {
	// load configuration from mineplus config file
	logMsh := config.LoadConfig()
	if logMsh != nil {
		logMsh.Log(true)
		progmgr.AutoTerminate()
	}

	// configure logging prefix after flags have been parsed by config.LoadConfig
	errco.ConfigureUserLogging(*logPrefix, *quiet)

	// launch wrapper manager
	go progmgr.MshMgr()
	// wait for the initial update check
	<-progmgr.ReqSent

	// if process suspension is allowed, pre-warm the server
	if config.ConfigRuntime.Msh.SuspendAllow {
		errco.NewLogln(errco.TYPE_INF, errco.LVL_1, errco.ERROR_NIL, "minecraft server will now pre-warm (process suspension is enabled)...")
		logMsh = servctrl.WarmMS()
		if logMsh != nil {
			logMsh.Log(true)
		}
	}

	// launch GetInput()
	go input.GetInput()

	// ---------------- connections ---------------- //

	// launch query handler
	if config.ConfigRuntime.Msh.EnableQuery {
		go conn.HandlerQuery()
	}

	// open a tcp listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.ProxyHost, config.ProxyPort))
	if err != nil {
		errco.NewLogln(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CLIENT_LISTEN, err.Error())
		progmgr.AutoTerminate()
	}

	// infinite cycle to handle new clients.
	errco.NewLogln(errco.TYPE_INF, errco.LVL_1, errco.ERROR_NIL, "%-40s %10s:%5d ...", "listening for new clients connections on", config.ProxyHost, config.ProxyPort)
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			errco.NewLogln(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CLIENT_ACCEPT, err.Error())
			continue
		}

		go conn.HandlerClientConn(clientConn)
	}
}
