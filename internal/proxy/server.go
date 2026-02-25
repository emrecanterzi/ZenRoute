package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/emrecanterzi/internal/dns"
)

type Server struct {
	addr         string
	resolver     dns.Resolver
	fragmentSize int
}

func NewServer(addr string, resolver dns.Resolver, fragmentSize int) *Server {
	return &Server{
		addr:         addr,
		resolver:     resolver,
		fragmentSize: fragmentSize,
	}
}

func (s *Server) Start(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("could not open port: %w", err)
	}
	defer listener.Close()

	fmt.Printf("proxy: listening on %s\n", s.addr)

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				fmt.Printf("proxy: accept error: %v\n", err)
				continue
			}
		}
		go s.handleConnection(clientConn)
	}
}

func (s *Server) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	buffer := make([]byte, 4096)
	data, err := clientConn.Read(buffer)
	if err != nil {
		return
	}

	requestString := string(buffer[:data])
	lines := strings.Split(requestString, "\n")
	if len(lines) == 0 {
		return
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) < 2 {
		return
	}

	target := parts[1]
	isDiscord := strings.Contains(target, "discord.com") || strings.Contains(target, "discord.gg")

	if !isDiscord {
		s.handleDirectTunnel(clientConn, target, parts[0], buffer[:data])
		return
	}

	s.handleSecureBypass(clientConn, target)
}

func (s *Server) handleDirectTunnel(clientConn net.Conn, target, method string, initialData []byte) {
	fmt.Printf("proxy: direct -> %s\n", target)
	serverConn, err := net.Dial("tcp", target)
	if err != nil {
		return
	}
	defer serverConn.Close()

	if method == "CONNECT" {
		clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		serverConn.Write(initialData)
	}

	s.bidirectionalCopy(clientConn, serverConn)
}

func (s *Server) handleSecureBypass(clientConn net.Conn, target string) {
	domain := target
	port := "443"

	if host, p, err := net.SplitHostPort(target); err == nil {
		domain = host
		port = p
	}

	realIP, err := s.resolver.Resolve(domain)
	if err != nil || realIP == "" {
		fmt.Printf("proxy: err resolving %s\n", domain)
		return
	}

	fmt.Printf("proxy: bypass -> %s (%s)\n", domain, realIP)
	serverConn, err := net.Dial("tcp", realIP+":"+port)
	if err != nil {
		return
	}
	defer serverConn.Close()

	if tcpConn, ok := serverConn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	// read TLS record header
	header := make([]byte, 5)
	_, err = io.ReadFull(clientConn, header)
	if err != nil {
		return
	}

	// 0x16 == TLS Handshake
	if header[0] == 0x16 {
		length := int(header[3])<<8 | int(header[4])

		body := make([]byte, length)
		_, err = io.ReadFull(clientConn, body)
		if err != nil {
			return
		}

		tlsData := append(header, body...)

		// shred ClientHello
		for i := 0; i < len(tlsData); i += s.fragmentSize {
			end := i + s.fragmentSize
			if end > len(tlsData) {
				end = len(tlsData)
			}
			serverConn.Write(tlsData[i:end])
			time.Sleep(2 * time.Millisecond)
		}
	} else {
		serverConn.Write(header)
	}

	s.bidirectionalCopy(clientConn, serverConn)
}

func (s *Server) bidirectionalCopy(clientConn, serverConn net.Conn) {
	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(serverConn, clientConn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, serverConn)
		errChan <- err
	}()

	<-errChan

	clientConn.Close()
	serverConn.Close()
}
