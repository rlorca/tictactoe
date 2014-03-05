package server

import ("net"
		"fmt"
		"log"
		"bufio"
)

type Server struct {
	listener net.Listener
}

type Command struct {
	payload string 	
}

func NewServer() *Server {
	return new(Server)
}

func (server *Server) Stop() {

	server.listener.Close()
}

func (server *Server) Start() {

	ln, err := net.Listen("tcp", ":2020")
	if err != nil {
		// handle error
	}

	server.listener = ln

	m := RunMatchMaker(Checker)

	for {
		
		conn, err := ln.Accept()

		if err != nil {
		
			log.Printf("Unable to accept new connections: %v", err)
			return

		} else {

			m.AddPlayer(NewPeer(conn));
		}
	}
}

type Peer struct {

	// the network connection
	conn net.Conn

	// this is how the peer sends messages to the server
	out chan Command

	// this is how the server communicates with the Peer
	in chan Command
}

func (p Peer) Perform(c *Command) {

	p.in <- *c
}

func (p Peer) quit() {

	p.conn.Close()

	close(p.out) 
	close(p.in)

	log.Printf("Done.")
}


func NewPeer(conn net.Conn) (*Peer) {

	//defer conn.Close()

	//TODO create a host abstraction here

	p := new(Peer)
	p.conn = conn
	p.out = make(chan Command)
	p.in = make(chan Command)

	go p.handleWrite()
	go p.handleRead()

	log.Printf("Connected! %v\n", conn)

	return p
}

func (p Peer) handleWrite() {

	w := bufio.NewWriter(p.conn)

	for {

		select {
			case c := <- p.in:
				fmt.Fprintf(w, c.payload)
				w.Flush()
		}
	}
}

func (p Peer) handleRead() {

	r := bufio.NewReader(p.conn)
	scanner := bufio.NewScanner(r)

	var c *Command

	for scanner.Scan() {

		t := scanner.Text()

		log.Printf("Handled %s %d\n", t, len(t))

		c = new(Command)
		c.payload = t
		p.out <- *c
	}

	if err := scanner.Err(); err != nil {

		log.Println("reading standard input:", err)

	} else {
		log.Println("no errors")
	}

	//p.quit()

	log.Printf("User %v left.", p.conn)
}