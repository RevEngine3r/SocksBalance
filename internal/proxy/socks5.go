package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// SOCKS5 protocol constants
const (
	socks5Version = 0x05

	// Authentication methods
	authNone     = 0x00
	authNoAccept = 0xFF

	// Commands
	cmdConnect = 0x01

	// Address types
	addrTypeIPv4   = 0x01
	addrTypeDomain = 0x03
	addrTypeIPv6   = 0x04

	// Reply codes
	replySuccess              = 0x00
	replyGeneralFailure       = 0x01
	replyConnectionNotAllowed = 0x02
	replyNetworkUnreachable   = 0x03
	replyHostUnreachable      = 0x04
	replyConnectionRefused    = 0x05
	replyTTLExpired           = 0x06
	replyCommandNotSupported  = 0x07
	replyAddrTypeNotSupported = 0x08
)

// handleSOCKS5 performs SOCKS5 handshake and returns target address
func handleSOCKS5(conn net.Conn) (string, error) {
	// Read version and methods
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", fmt.Errorf("failed to read version: %w", err)
	}

	version := buf[0]
	nMethods := buf[1]

	if version != socks5Version {
		return "", fmt.Errorf("unsupported SOCKS version: %d", version)
	}

	// Read authentication methods
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return "", fmt.Errorf("failed to read methods: %w", err)
	}

	// Check if NO_AUTH is supported
	hasNoAuth := false
	for _, m := range methods {
		if m == authNone {
			hasNoAuth = true
			break
		}
	}

	if !hasNoAuth {
		// Send no acceptable methods
		conn.Write([]byte{socks5Version, authNoAccept})
		return "", fmt.Errorf("no supported authentication method")
	}

	// Send NO_AUTH selected
	if _, err := conn.Write([]byte{socks5Version, authNone}); err != nil {
		return "", fmt.Errorf("failed to send auth response: %w", err)
	}

	// Read request
	reqHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, reqHeader); err != nil {
		return "", fmt.Errorf("failed to read request header: %w", err)
	}

	if reqHeader[0] != socks5Version {
		return "", fmt.Errorf("invalid request version: %d", reqHeader[0])
	}

	cmd := reqHeader[1]
	addrType := reqHeader[3]

	// Only support CONNECT
	if cmd != cmdConnect {
		sendReply(conn, replyCommandNotSupported)
		return "", fmt.Errorf("unsupported command: %d", cmd)
	}

	// Read address
	var addr string
	switch addrType {
	case addrTypeIPv4:
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return "", fmt.Errorf("failed to read IPv4: %w", err)
		}
		addr = net.IP(ip).String()

	case addrTypeDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return "", fmt.Errorf("failed to read domain length: %w", err)
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", fmt.Errorf("failed to read domain: %w", err)
		}
		addr = string(domain)

	case addrTypeIPv6:
		ip := make([]byte, 16)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return "", fmt.Errorf("failed to read IPv6: %w", err)
		}
		addr = net.IP(ip).String()

	default:
		sendReply(conn, replyAddrTypeNotSupported)
		return "", fmt.Errorf("unsupported address type: %d", addrType)
	}

	// Read port
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return "", fmt.Errorf("failed to read port: %w", err)
	}
	port := binary.BigEndian.Uint16(portBuf)

	target := fmt.Sprintf("%s:%d", addr, port)

	// Send success reply
	if err := sendReply(conn, replySuccess); err != nil {
		return "", fmt.Errorf("failed to send reply: %w", err)
	}

	return target, nil
}

// sendReply sends a SOCKS5 reply to the client
func sendReply(conn net.Conn, reply byte) error {
	// Build reply: VER REP RSV ATYP ADDR PORT
	// Using 0.0.0.0:0 as bind address
	resp := []byte{
		socks5Version,
		reply,
		0x00, // Reserved
		addrTypeIPv4,
		0, 0, 0, 0, // 0.0.0.0
		0, 0, // Port 0
	}

	_, err := conn.Write(resp)
	return err
}
