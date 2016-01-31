package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
    "os"
    "time"
)

/*
    Copyright 2016 Alexander Scheel <alexander.m.scheel@gmail.com>
    Licensed under the GPLv3

    ~~

    go-sshlog sits and logs incoming ssh attempts on a given port with a given ssh key
*/

func main() {
    if (len(os.Args) == 2 && os.Args[1] == "help") || len(os.Args) > 3 {
        fmt.Println("Usage:", os.Args[0], "[port] [private_key]\nCopyright (C) 2016 Alexander Scheel <alexander.m.scheel@gmail.com>\nLicensed under the GPLv3\n\n\t[port]\t\t- connection string to listen on; default is 0.0.0.0:22\n\t[private_key]\t- private key file for ssh server; default is ./id_rsa\n\nTo generate a new SSH key:\n\tssh-keygen -t rsa -f id_rsa")
        os.Exit(0)
    }

    port := "0.0.0.0:22"
    private_key := "id_rsa"

    for i := range os.Args {
        if i == 0 {
            continue
        }

        // If it exists, it's  path; otherwise assume it is a port
        if _, err := os.Stat(os.Args[i]); os.IsNotExist(err) {
            port = os.Args[i]
        } else {
            private_key = os.Args[i]
        }
    }

	config := &ssh.ServerConfig {
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            // Log and deny
			fmt.Println("time:", time.Now().Format(time.RFC3339), "| from:", c.RemoteAddr().String(), "| user:", string(c.User()), "| pass:", string(pass))
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := ioutil.ReadFile(private_key)
	if err != nil {
		log.Fatal("Failed to load private key (./id_rsa)")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("Failed to listen on %s (%s)", port, err)
	}

	// Accept all connections
	fmt.Println("Listening...")
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		_, _, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}
		go ssh.DiscardRequests(reqs)
	}
}
