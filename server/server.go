package server

import (
	"fmt"
	"net"
	"simpleKV/resp"
	"simpleKV/server/store"
)

type IServer interface {
	Run() error
}

type server struct {
	store store.IStore
	addr  string
}

func NewServer(addr string, store store.IStore) IServer {
	return &server{
		store: store,
		addr:  addr,
	}
}

func (s *server) Run() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("Error starting server: %v", err)
	}
	defer ln.Close()

	fmt.Printf("Server started on %s\n", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := resp.NewReader(conn)

	for {
		req, err := reader.Read()
		if err != nil {
			fmt.Println("Error reading request:", err)
			return
		}

		res := s.handleRequest(req)
		_, err = conn.Write(res.Marshal())
		if err != nil {
			fmt.Println("Error writing response:", err)
			return
		}
	}
}
