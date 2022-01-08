// Websocket utility functions
//
// Copyright (c) 2021  Philip Kaludercic
//
// This file is part of go-kgp.
//
// go-kgp is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License,
// version 3, as published by the Free Software Foundation.
//
// go-kgp is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License, version 3, along with go-kgp. If not, see
// <http://www.gnu.org/licenses/>

package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func listenUpgrade(w http.ResponseWriter, r *http.Request) {
	// upgrade to websocket or bail out
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Unable to upgrade connection: %s", err)
		w.WriteHeader(400)
		return
	}
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())
	(&Client{rwc: &wsrwc{Conn: conn}}).Handle()
}

// adapted from https://github.com/gorilla/websocket/issues/282

// wsrwc is a read-write-closer using websockets
type wsrwc struct {
	*websocket.Conn
	r io.Reader
}

func (c *wsrwc) Write(p []byte) (int, error) {
	err := c.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *wsrwc) Read(p []byte) (int, error) {
	for {
		if c.r == nil {
			// Advance to next message.
			var err error
			_, c.r, err = c.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.r.Read(p)
		if err == io.EOF {
			// At end of message.
			c.r = nil
			if n > 0 {
				return n, nil
			} else {
				// No data read, continue to next message.
				continue
			}
		}
		return n, err
	}
}
