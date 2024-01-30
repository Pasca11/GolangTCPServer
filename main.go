package main

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	from string
	msg  []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitChan   chan struct{}
	msgChan    chan Message
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitChan:   make(chan struct{}),
		msgChan:    make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.acceptLoop()

	<-s.quitChan
	close(s.msgChan)
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error", err)
			continue
		}
		fmt.Println("new connection to", conn.RemoteAddr())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error", err)
			continue
		}
		s.msgChan <- Message{
			from: conn.RemoteAddr().String(),
			msg:  buf[:n],
		}

		conn.Write([]byte("Got it!\n"))
	}
}

func main() {
	server := NewServer(":3000")
	go func() {
		for msg := range server.msgChan {
			fmt.Printf("recieved message from %s: %s", msg.from, string(msg.msg))
		}
	}()

	log.Fatalln(server.Start())
}
