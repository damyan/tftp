package tftp

import (
	"net"
	"fmt"
	"io"
	"log"
)

/*
Server provides TFTP server functionality. It requires handlers bind address,
handlers for read and write requests and optional logger.

	func HandleWrite(filename string, r *io.PipeReader) {
		buffer := &bytes.Buffer{}
		buffer.ReadFrom(r)
		...
	}
	func HandleRead(filename string, w *io.PipeWriter) {
		if fileExists {
			...
			buffer.WriteTo(w)
			w.Close()
		} else {
			w.CloseWithError(fmt.Errorf("File not exists: %s", filename))
		}
	}
	...
	addr, e := net.ResolveUDPAddr("udp", ":69")
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	log := log.New(os.Stderr, "TFTP", log.Ldate | log.Ltime)
	s := tftp.Server{addr, HandleWrite, HandleRead, log}
	e = s.Serve()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
*/
type Server struct {
	BindAddr *net.UDPAddr
	ReadHandler func(filename string, r *io.PipeReader)
	WriteHandler func(filename string, w *io.PipeWriter)
	Log *log.Logger
}

func (s Server) Serve() (error) {
	conn, e := net.ListenUDP("udp", s.BindAddr)
	if e != nil {
		return e
	}
	for {
		e = s.processRequest(conn)
		if e != nil {
			if s.Log != nil {
				s.Log.Printf("%v\n", e);
			}
		}
	}
}

func (s Server) processRequest(conn *net.UDPConn) (error) {
	var buffer []byte
	buffer = make([]byte, 50)
	
	written, addr, e := conn.ReadFromUDP(buffer)
	if e != nil {
		return fmt.Errorf("Failed to read data from client: %v", e)
	}
	p, e := ParsePacket(buffer[:written])
	
	switch p := Packet(*p).(type) {
		case *WRQ:
			trasnmissionConn, e := s.transmissionConn()
			if e != nil {
				return fmt.Errorf("Could not start transmission: %v", e)
			}
			reader, writer := io.Pipe()
			r := &receiver{addr, trasnmissionConn, writer, s.Log}
			go s.ReadHandler(p.Filename, reader)
			go r.Run()
		case *RRQ:
			trasnmissionConn, e := s.transmissionConn()
			if e != nil {
				return fmt.Errorf("Could not start transmission: %v", e)
			}
			reader, writer := io.Pipe()
			r := &sender{addr, trasnmissionConn, reader, s.Log}
			go s.WriteHandler(p.Filename, writer)
			go r.Run()
	}
	return nil
}

func (s Server) transmissionConn() (*net.UDPConn, error) {
	addr, e := net.ResolveUDPAddr("udp", ":0")
	if e != nil {
		return nil, e
	}
	conn, e := net.ListenUDP("udp", addr)
	if e != nil {
		return nil, e
	}
	return conn, nil
}