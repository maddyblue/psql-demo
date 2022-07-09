// psql-demo is a program for handling postgres connections and replying with a single NOTICE response.
//
// Configuration via env vars:
//
// PSQLDEMO_ADDR: listen address, default "0.0.0.0:5432"
//
// PSQLDEMO_MESSAGE: notice response text. Can contain "\n" for newlines.
package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgproto3/v2"
	"github.com/kelseyhightower/envconfig"
)

type Specification struct {
	Addr    string `default:"0.0.0.0:5432"`
	Message string `default:"https://www.youtube.com/watch?v=dQw4w9WgXcQ"`
}

func main() {
	var s Specification
	err := envconfig.Process("psqldemo", &s)
	if err != nil {
		panic(err)
	}
	s.Message = strings.ReplaceAll(s.Message, "\\n", "\n")

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Println("serving on", s.Addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Specification) handleConn(conn net.Conn) {
	defer conn.Close()
	err := conn.SetDeadline(time.Now().Add(time.Second * 10))
	if err != nil {
		fmt.Println(err)
		return
	}

	backend := pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn)
	for {
		msg, err := backend.ReceiveStartupMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		if _, ok := msg.(*pgproto3.SSLRequest); ok {
			// Say no to SSL.
			if _, err := io.WriteString(conn, "N"); err != nil {
				fmt.Println(err)
				return
			}
		} else {
			break
		}
	}
	err = backend.Send(&pgproto3.AuthenticationOk{})
	if err != nil {
		fmt.Println(err)
		return
	}
	err = backend.Send(&pgproto3.NoticeResponse{
		Severity: "NOTICE",
		Message:  s.Message,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Drain any other messages. Don't print errors here because we expect them.
	for {
		msg, err := backend.Receive()
		if err != nil {
			break
		}
		if _, ok := msg.(*pgproto3.Terminate); ok {
			break
		}
	}
}
