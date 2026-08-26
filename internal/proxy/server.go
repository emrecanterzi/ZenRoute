package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/emrecanterzi/zenroute/internal/dns"
	"github.com/emrecanterzi/zenroute/internal/util"
)

type Options struct {
	Addr          string
	FragmentSize  int
	BypassDomains []string
	BypassAll     bool
}

type Server struct {
	opts     Options
	resolver dns.Resolver
	hostname string
	port     string
}

func NewServer(opts Options, resolver dns.Resolver) *Server {
	hostname := "localhost"
	if h, err := os.Hostname(); err != nil {
		fmt.Printf("zenroute: could not get hostname, falling back to localhost %v\n", err)
	} else {
		hostname = h
	}

	port := "8080"
	_, p, err := net.SplitHostPort(opts.Addr)
	if err == nil {
		port = p
	}

	return &Server{
		opts:     opts,
		resolver: resolver,
		hostname: hostname,
		port:     port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.opts.Addr)
	if err != nil {
		return fmt.Errorf("could not open port: %w", err)
	}
	defer listener.Close()

	fmt.Printf("proxy: listening on %s\n", s.opts.Addr)

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
	n, err := clientConn.Read(buffer)
	if err != nil {
		return
	}

	requestString := string(buffer[:n])
	lines := strings.Split(requestString, "\n")
	if len(lines) == 0 {
		return
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) < 2 {
		return
	}

	method := parts[0]
	target := parts[1]

	if method == "GET" && strings.HasPrefix(target, "/proxy.pac") {
		clientConn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/x-ns-proxy-autoconfig\r\n\r\nfunction FindProxyForURL(url, host) {\n  return \"PROXY " + s.hostname + ".local:" + s.port + "\";\n}"))
		return
	}

	shouldBypass := s.shouldBypass(method, target)

	if !shouldBypass {
		s.handleDirectTunnel(clientConn, target, parts[0], buffer[:n])
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
		domain = util.NormalizeDomain(host)
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

		// shred ClientHello
		for i := 0; i < len(body); i += s.opts.FragmentSize {
			end := i + s.opts.FragmentSize
			if end > len(body) {
				end = len(body)
			}
			fragment := body[i:end]
			record := make([]byte, 5+len(fragment))
			record[0] = 0x16
			record[1] = header[1]
			record[2] = header[2]
			record[3] = byte(len(fragment) >> 8)
			record[4] = byte(len(fragment))
			copy(record[5:], fragment)
			serverConn.Write(record)
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

func (s *Server) shouldBypass(method, target string) bool {
	if method != "CONNECT" {
		return false
	}

	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return false
	}

	host = util.NormalizeDomain(host)

	if s.opts.BypassAll {
		return true
	}

	for _, domain := range s.opts.BypassDomains {
		domain = util.NormalizeDomain(domain)

		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}
