package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/emrecanterzi/zenroute/internal/dns"
	"github.com/emrecanterzi/zenroute/internal/util"
)

type Options struct {
	Addr          string
	BypassDomains []string
	BypassAll     bool
	MaxConns      int
}

type Server struct {
	opts     Options
	resolver dns.Resolver
	hostname string
	port     string
}

func NewServer(opts Options, resolver dns.Resolver) *Server {
	hostname := ""
	if h, err := os.Hostname(); err != nil {
		fmt.Printf("zenroute: could not get hostname, /proxy.pac will be unavailable: %v\n", err)
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

	sem := make(chan struct{}, s.opts.MaxConns)

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
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			clientConn.Close()
			return nil
		}
		go func() {
			defer func() { <-sem }()
			s.handleConnection(clientConn)
		}()
	}
}

func (s *Server) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	buffer := make([]byte, 4096)
	clientConn.SetReadDeadline(time.Now().Add(time.Second * 5))
	n, err := clientConn.Read(buffer)
	if err != nil {
		return
	}
	clientConn.SetReadDeadline(time.Time{})

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
		if s.hostname == "" {
			clientConn.Write([]byte("HTTP/1.1 503 Service Unavailable\r\nContent-Length: 0\r\n\r\n"))
			return
		}
		clientConn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/x-ns-proxy-autoconfig\r\n\r\nfunction FindProxyForURL(url, host) {\n  return \"PROXY " + s.hostname + ":" + s.port + "; DIRECT\";\n}"))
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

	dialAddr := target
	if method != "CONNECT" {
		u, err := url.Parse(target)
		if err != nil || u.Host == "" {
			return
		}
		dialAddr = u.Host
		if u.Port() == "" {
			dialAddr = net.JoinHostPort(u.Hostname(), "80")
		}
	}

	serverConn, err := net.DialTimeout("tcp", dialAddr, 10*time.Second)
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

	serverConn, err := net.DialTimeout("tcp", realIP+":"+port, 10*time.Second)
	if err != nil {
		return
	}
	defer serverConn.Close()

	if tcpConn, ok := serverConn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	header := make([]byte, 5)
	clientConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = io.ReadFull(clientConn, header)
	if err != nil {
		return
	}

	if header[0] == 0x16 {
		length := int(header[3])<<8 | int(header[4])
		body := make([]byte, length)
		_, err = io.ReadFull(clientConn, body)
		if err != nil {
			return
		}

		sniStart, sniLen, err := s.findSNIRange(body)
		if err != nil {
			fmt.Printf("proxy: sni parse failed for %s: %v\n", domain, err)
			return
		}

		fmt.Printf("proxy: fragmented -> %s (%s)\n", domain, realIP)

		splitAt := sniStart + sniLen/2
		for _, chunk := range [][]byte{body[:splitAt], body[splitAt:]} {
			record := make([]byte, 5+len(chunk))
			record[0] = 0x16
			record[1], record[2] = header[1], header[2]
			record[3] = byte(len(chunk) >> 8)
			record[4] = byte(len(chunk))
			copy(record[5:], chunk)
			serverConn.Write(record)
			time.Sleep(2 * time.Millisecond)
		}
	} else {
		fmt.Printf("proxy: raw passthrough (not TLS handshake) -> %s (%s)\n", domain, realIP)
		serverConn.Write(header)
	}
	clientConn.SetReadDeadline(time.Time{})
	s.bidirectionalCopy(clientConn, serverConn)
}

func (s *Server) bidirectionalCopy(clientConn, serverConn net.Conn) {
	errChan := make(chan error, 2)

	for _, c := range []net.Conn{clientConn, serverConn} {
		if tcpConn, ok := c.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}
	}

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
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

func (s *Server) findSNIRange(body []byte) (start, length int, err error) {
	pos := 4 + 2 + 32

	if pos+1 > len(body) {
		return 0, 0, fmt.Errorf("truncated before session id")
	}
	pos += 1 + int(body[pos])

	if pos+2 > len(body) {
		return 0, 0, fmt.Errorf("truncated before cipher suites")
	}

	csLen := int(body[pos])<<8 | int(body[pos+1])
	pos += 2 + csLen

	if pos+1 > len(body) {
		return 0, 0, fmt.Errorf("truncated before compression methods")
	}

	pos += 1 + int(body[pos])

	if pos+2 > len(body) {
		return 0, 0, fmt.Errorf("no extensions")
	}

	extsLen := int(body[pos])<<8 | int(body[pos+1])
	pos += 2
	extsEnd := pos + extsLen
	if extsEnd > len(body) {
		return 0, 0, fmt.Errorf("extensions length overruns body")
	}

	for pos+4 <= extsEnd {
		extType := int(body[pos])<<8 | int(body[pos+1])
		extLen := int(body[pos+2])<<8 | int(body[pos+3])
		dataStart := pos + 4

		if extType == 0x0000 {
			if extLen < 5 || dataStart+extLen > len(body) {
				return 0, 0, fmt.Errorf("malformed sni extension")
			}
			if body[dataStart+2] != 0x00 {
				return 0, 0, fmt.Errorf("unsupported sni name type")
			}
			nameLen := int(body[dataStart+3])<<8 | int(body[dataStart+4])
			nameStart := dataStart + 5
			if nameStart+nameLen > len(body) {
				return 0, 0, fmt.Errorf("sni name overruns body")
			}
			return nameStart, nameLen, nil
		}
		pos = dataStart + extLen
	}

	return 0, 0, fmt.Errorf("sni extension not found")
}
