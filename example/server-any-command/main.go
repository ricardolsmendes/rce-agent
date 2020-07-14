// Copyright 2020 Square, Inc.

/*
	This is an example RCE server (agent). It uses an rce.Server (which uses a
	gRPC server) to run any command requested by clients.

	NOTE: the exec commands must be absolute paths! "/bin/ls" not "ls"!

	This example code and your agent code should be similar because there is not
	much variation for running the server. One thing that will be different:
	to make this agent long-running, we purposely block on channel recv and wait
	for a CTRL-C signal to gracefully shutdown. Your agent might be in an API
	or other back end system that's inherently long-running.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/square/rce-agent"
)

var (
	flagTLSCert         string
	flagTLSKey          string
	flagTLSCA           string
	flagAddr            string
	flagAllowAnyCommand bool
	flagDisableSecurity bool
)

func init() {
	flag.StringVar(&flagTLSCert, "tls-cert", "", "TLS certificate file")
	flag.StringVar(&flagTLSKey, "tls-key", "", "TLS key file")
	flag.StringVar(&flagTLSCA, "tls-ca", "", "TLS certificate authority")
	flag.StringVar(&flagAddr, "addr", "127.0.0.1:5501", "Address and port to listen on")
	flag.BoolVar(&flagAllowAnyCommand, "allow-any-command", false, "Allow any command")
	flag.BoolVar(&flagDisableSecurity, "disable-security", false, "Disable security")
}

func main() {
	// ----------------------------------------------------------------------
	// Parse command line flags (options)
	// ----------------------------------------------------------------------
	flag.Parse()

	// ----------------------------------------------------------------------
	// Load TLS if given
	// ----------------------------------------------------------------------
	// You should use rce.TLSFiles like used here because it creates a
	// tls.Config that requires mutual authentication: client verifies agent
	// TLS cert _and_ agent verifies client TLS cert. You can create your
	// own tls.Config if you don't need mutual auth.
	tlsFiles := rce.TLSFiles{
		CACert: flagTLSCA,
		Cert:   flagTLSCert,
		Key:    flagTLSKey,
	}
	tlsConfig, err := tlsFiles.TLSConfig()
	if err != nil {
		log.Fatal(err)
	}
	if tlsConfig != nil {
		log.Println("TLS loaded")
	}

	// ----------------------------------------------------------------------
	// Create a ServerConfig object
	// ----------------------------------------------------------------------
	cfg := &rce.ServerConfig{
		Addr:            flagAddr,
		AllowAnyCommand: flagAllowAnyCommand,
		DisableSecurity: flagDisableSecurity,
		TLS:             tlsConfig,
	}

	// ----------------------------------------------------------------------
	// Create and start agent
	// ----------------------------------------------------------------------
	srv := rce.NewServerWithConfig(*cfg)
	if err := srv.StartServer(); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}

	// As the docs say, StartServer() is non-blocking, so the server is
	// listening for client connections at this point. Everything else is
	// handled internally, nothing else to do to make clients work. We just
	// need to keep this Go program running...

	// ----------------------------------------------------------------------
	// Wait for CTRL-C for graceful shutdown
	// ----------------------------------------------------------------------
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	fmt.Println("CTRL-C to shut down")
	<-c
	fmt.Println("Shutting down...")
	if err := srv.StopServer(); err != nil {
		log.Printf("Error stopping server: %s\n", err)
	}
}
