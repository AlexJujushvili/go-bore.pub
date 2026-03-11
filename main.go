package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	boreServer     = "bore.pub"
	localPort      = uint16(8080)
	remotePort     = uint16(0) // 0 = random
	controlPort    = 7835
	networkTimeout = 10 * time.Second
	maxFrameLen    = 256
)

// Delimiter: null byte \0 (AnyDelimiterCodec)
// Framing: JSON + \0

func writeMsg(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, 0) // null byte delimiter
	_, err = w.Write(data)
	return err
}

func sendHello(w io.Writer, port uint16) error {
	return writeMsg(w, map[string]any{"Hello": port})
}

func sendAccept(w io.Writer, uuid string) error {
	return writeMsg(w, map[string]any{"Accept": uuid})
}

// readFrame reads until null byte \0
func readFrame(conn net.Conn) ([]byte, error) {
	var buf []byte
	b := make([]byte, 1)
	for {
		conn.SetReadDeadline(time.Now().Add(networkTimeout))
		_, err := conn.Read(b)
		if err != nil {
			return nil, err
		}
		if b[0] == 0 {
			return buf, nil
		}
		buf = append(buf, b[0])
		if len(buf) > maxFrameLen {
			return nil, fmt.Errorf("frame too long")
		}
	}
}

type ServerMsg struct {
	Hello      *uint16
	Connection *string
	Challenge  *string
	Error      *string
	Heartbeat  bool
}

func recvServerMsg(conn net.Conn) (*ServerMsg, error) {
	frame, err := readFrame(conn)
	if err != nil {
		return nil, err
	}

	// unit variant: "Heartbeat"
	var s string
	if json.Unmarshal(frame, &s) == nil {
		if s == "Heartbeat" {
			return &ServerMsg{Heartbeat: true}, nil
		}
		return nil, fmt.Errorf("unknown string message: %s", s)
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(frame, &obj); err != nil {
		return nil, fmt.Errorf("parse error: %w (raw: %s)", err, frame)
	}

	msg := &ServerMsg{}
	if v, ok := obj["Hello"]; ok {
		var port uint16
		json.Unmarshal(v, &port)
		msg.Hello = &port
	}
	if v, ok := obj["Connection"]; ok {
		var uuid string
		json.Unmarshal(v, &uuid)
		msg.Connection = &uuid
	}
	if v, ok := obj["Challenge"]; ok {
		var uuid string
		json.Unmarshal(v, &uuid)
		msg.Challenge = &uuid
	}
	if v, ok := obj["Error"]; ok {
		var errMsg string
		json.Unmarshal(v, &errMsg)
		msg.Error = &errMsg
	}
	return msg, nil
}

func dialServer() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", boreServer, controlPort)
	return net.DialTimeout("tcp", addr, networkTimeout)
}

func handleConnection(uuid, localAddr string) {
	remote, err := dialServer()
	if err != nil {
		fmt.Printf("❌ server re-dial failed: %v\n", err)
		return
	}
	defer remote.Close()

	if err := sendAccept(remote, uuid); err != nil {
		fmt.Printf("❌ accept send failed: %v\n", err)
		return
	}

	local, err := net.DialTimeout("tcp", localAddr, networkTimeout)
	if err != nil {
		fmt.Printf("❌ local dial failed (%s): %v\n", localAddr, err)
		return
	}
	defer local.Close()

	// read_buf flush — remaining bytes after framing (like Rust's into_parts)
	// bore სერვერი Accept-ის შემდეგ raw TCP-ზე გადადის,
	// ამიტომ buffered data-ც უნდა გავუშვათ local-ზე
	done := make(chan struct{}, 2)
	go func() { io.Copy(remote, local); done <- struct{}{} }()
	go func() { io.Copy(local, remote); done <- struct{}{} }()
	<-done
}

// bufConn — net.Conn wrapper with prepended buffer (for leftover bytes)
type bufConn struct {
	net.Conn
	buf *bytes.Reader
}

func (b *bufConn) Read(p []byte) (int, error) {
	if b.buf.Len() > 0 {
		return b.buf.Read(p)
	}
	return b.Conn.Read(p)
}

func run(ctx context.Context) error {
	ctrl, err := dialServer()
	if err != nil {
		return fmt.Errorf("cannot connect to %s:%d: %w", boreServer, controlPort, err)
	}
	defer ctrl.Close()

	fmt.Printf("🔌 connected to %s:%d\n", boreServer, controlPort)

	// bore.pub has no secret → no Challenge → directly send Hello
	if err := sendHello(ctrl, remotePort); err != nil {
		return fmt.Errorf("send hello: %w", err)
	}

	msg, err := recvServerMsg(ctrl)
	if err != nil {
		return fmt.Errorf("handshake: %w", err)
	}

	if msg.Error != nil {
		return fmt.Errorf("server error: %s", *msg.Error)
	}
	if msg.Hello == nil {
		return fmt.Errorf("unexpected handshake response: %+v", msg)
	}

	assigned := *msg.Hello
	localAddr := fmt.Sprintf("localhost:%d", localPort)
	fmt.Printf("✅ tunnel open!\n")
	fmt.Printf("   remote → %s:%d\n", boreServer, assigned)
	fmt.Printf("   local  → %s\n\n", localAddr)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := recvServerMsg(ctrl)
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("server closed connection")
			}
			return err
		}

		if msg.Heartbeat {
			continue
		}
		if msg.Error != nil {
			return fmt.Errorf("server: %s", *msg.Error)
		}
		if msg.Connection != nil {
			go handleConnection(*msg.Connection, localAddr)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; fmt.Println("\n🛑 closing..."); cancel() }()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
