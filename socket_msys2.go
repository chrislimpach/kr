// +build !darwin

package kr

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"time"
)

type AgentListener struct {
	net.Listener
	secret [4]uint32
	peer_cred [3]uint32
}

type AgentConn struct {
	net.Conn
}

func recvUint32(l net.Conn) (v uint32, err error) {
	var buf [4]byte
	var n int

	n, err = l.Read(buf[:])
	if err != nil {
		return
	}
	if n != 4 {
		err = io.ErrUnexpectedEOF
		return
	}
	v = binary.LittleEndian.Uint32(buf[:])
	return
}

func sendUint32(l net.Conn, v uint32) (err error) {
	var buf [4]byte
	var n int

	binary.LittleEndian.PutUint32(buf[:], v)

	n, err = l.Write(buf[:])
	if err == nil && n != 4 {
		err = io.ErrUnexpectedEOF
	}
	return
}

func newAgentListener(path string) (s AgentListener, err error) {
	var l net.Listener

	l, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		return
	}

	// log.Error(fmt.Sprintf("Server listening on [%s]", l.Addr()))

	s.secret[0] = rand.Uint32()
	s.secret[1] = rand.Uint32()
	s.secret[2] = rand.Uint32()
	s.secret[3] = rand.Uint32()
	sid := []byte(fmt.Sprintf("!<socket >%d s %08X-%08X-%08X-%08X",
		l.Addr().(*net.TCPAddr).Port,
		s.secret[0], s.secret[1], s.secret[2], s.secret[3]))
	// log.Error(string(sid))

	err = ioutil.WriteFile(path, sid, 0600)
	if err != nil {
		l.Close()
		return
	}

	exec.Command("chattr", "+s", path).Run()
	<-time.After(1*time.Second)

	s.Listener = l
	return
}

func (s *AgentListener) Accept() (c AgentConn, err error) {
	var nc net.Conn

	nc, err = s.Listener.Accept()
	if err != nil {
		return
	}

	c.Conn = nc
	return
}

func msysRecvUint32(l net.Conn) (v uint32, err error) {
	var buf [4]byte
	var n int

	n, err = l.Read(buf[:])
	if err != nil {
		return
	}
	if n != 4 {
		err = io.ErrUnexpectedEOF
		return
	}
	v = binary.LittleEndian.Uint32(buf[:])
	return
}

func msysSendUint32(l net.Conn, v uint32) (err error) {
	var buf [4]byte
	var n int

	binary.LittleEndian.PutUint32(buf[:], v)

	n, err = l.Write(buf[:])
	if err == nil && n != 4 {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (c *AgentConn) Handshake(s AgentListener) (err error) {
	var in_secret [4]uint32
	var in_cred [3]uint32

	// log.Notice(fmt.Sprintf("Client connected [%s]", c.RemoteAddr()))

	for n, _ := range in_secret {
		in_secret[n], err = msysRecvUint32(c)
		if err != nil {
			// log.Error("msysRecvUint32 in_secret[", n, "]")
			return
		}
	}
	// log.Notice(fmt.Sprintf("secret: %08X-%08X-%08X-%08X",
	// 	in_secret[0], in_secret[1], in_secret[2], in_secret[3]))

	if s.secret[0] != in_secret[0] || s.secret[1] != in_secret[1] ||
		s.secret[2] != in_secret[2] || s.secret[2] != in_secret[2] {
		// log.Error("unix socket secret mismatch")
		err = errors.New("unix socket secret mismatch")
		return
	}

	for _, v := range in_secret {
		err = msysSendUint32(c, v)
		if err != nil {
			// log.Error("msysSendUint32 in_secret[", n, "]")
			return
		}
	}

	for n, _ := range s.peer_cred {
		s.peer_cred[n], err = msysRecvUint32(c)
		if err != nil {
			// log.Error("msysRecvUint32 s.peer_cred[", n, "]")
			return
		}
	}

	// log.Notice(fmt.Sprintf("cred: %08X-%08X-%08X",
	// 	s.peer_cred[0], s.peer_cred[1], s.peer_cred[2]))

	in_cred[0] = uint32(os.Getpid())
	in_cred[1] = uint32(os.Getuid())
	in_cred[2] = uint32(os.Getgid())

	for _, v := range in_cred {
		err = msysSendUint32(c, v)
		if err != nil {
			// log.Error("msysSendUint32 in_cred[", n, "]")
			return
		}
	}

	// log.Notice(fmt.Sprintf("Client handshake done [%s]", c.RemoteAddr()))
	return
}
