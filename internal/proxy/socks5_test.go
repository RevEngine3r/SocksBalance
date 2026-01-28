package proxy

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  bytes.NewBuffer(nil),
		writeBuf: bytes.NewBuffer(nil),
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestHandleSOCKS5ConnectIPv4(t *testing.T) {
	conn := newMockConn()

	// Client request: version, nMethods, methods
	conn.readBuf.Write([]byte{socks5Version, 1, authNone})

	// Client CONNECT request: IPv4
	req := []byte{
		socks5Version,
		cmdConnect,
		0x00, // Reserved
		addrTypeIPv4,
		8, 8, 8, 8, // 8.8.8.8
	}
	port := uint16(53)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, port)
	req = append(req, portBytes...)
	conn.readBuf.Write(req)

	target, err := handleSOCKS5(conn)
	if err != nil {
		t.Fatalf("handleSOCKS5 failed: %v", err)
	}

	expected := "8.8.8.8:53"
	if target != expected {
		t.Errorf("Expected target %s, got %s", expected, target)
	}

	// Verify auth response
	written := conn.writeBuf.Bytes()
	if len(written) < 2 {
		t.Fatal("No auth response written")
	}
	if written[0] != socks5Version || written[1] != authNone {
		t.Error("Invalid auth response")
	}
}

func TestHandleSOCKS5ConnectDomain(t *testing.T) {
	conn := newMockConn()

	// Client request
	conn.readBuf.Write([]byte{socks5Version, 1, authNone})

	// CONNECT request with domain
	domain := "example.com"
	req := []byte{
		socks5Version,
		cmdConnect,
		0x00,
		addrTypeDomain,
		byte(len(domain)),
	}
	req = append(req, []byte(domain)...)
	port := uint16(80)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, port)
	req = append(req, portBytes...)
	conn.readBuf.Write(req)

	target, err := handleSOCKS5(conn)
	if err != nil {
		t.Fatalf("handleSOCKS5 failed: %v", err)
	}

	expected := "example.com:80"
	if target != expected {
		t.Errorf("Expected target %s, got %s", expected, target)
	}
}

func TestHandleSOCKS5InvalidVersion(t *testing.T) {
	conn := newMockConn()

	// Invalid version
	conn.readBuf.Write([]byte{0x04, 1, authNone})

	_, err := handleSOCKS5(conn)
	if err == nil {
		t.Error("Expected error for invalid version")
	}
}

func TestHandleSOCKS5NoAuthSupported(t *testing.T) {
	conn := newMockConn()

	// No NO_AUTH method
	conn.readBuf.Write([]byte{socks5Version, 1, 0x02})

	_, err := handleSOCKS5(conn)
	if err == nil {
		t.Error("Expected error when NO_AUTH not supported")
	}

	// Should send authNoAccept
	written := conn.writeBuf.Bytes()
	if len(written) >= 2 && written[1] != authNoAccept {
		t.Error("Expected authNoAccept response")
	}
}

func TestHandleSOCKS5UnsupportedCommand(t *testing.T) {
	conn := newMockConn()

	conn.readBuf.Write([]byte{socks5Version, 1, authNone})

	// BIND command (not supported)
	req := []byte{
		socks5Version,
		0x02, // BIND
		0x00,
		addrTypeIPv4,
		8, 8, 8, 8,
		0, 80,
	}
	conn.readBuf.Write(req)

	_, err := handleSOCKS5(conn)
	if err == nil {
		t.Error("Expected error for unsupported command")
	}
}

func TestHandleSOCKS5UnsupportedAddrType(t *testing.T) {
	conn := newMockConn()

	conn.readBuf.Write([]byte{socks5Version, 1, authNone})

	// Invalid address type
	req := []byte{
		socks5Version,
		cmdConnect,
		0x00,
		0xFF, // Invalid
		8, 8, 8, 8,
		0, 80,
	}
	conn.readBuf.Write(req)

	_, err := handleSOCKS5(conn)
	if err == nil {
		t.Error("Expected error for unsupported address type")
	}
}

func TestSendReply(t *testing.T) {
	conn := newMockConn()

	err := sendReply(conn, replySuccess)
	if err != nil {
		t.Fatalf("sendReply failed: %v", err)
	}

	written := conn.writeBuf.Bytes()
	if len(written) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(written))
	}

	if written[0] != socks5Version {
		t.Error("Invalid version in reply")
	}
	if written[1] != replySuccess {
		t.Error("Invalid reply code")
	}
	if written[3] != addrTypeIPv4 {
		t.Error("Invalid address type in reply")
	}
}
