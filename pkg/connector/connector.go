package connector

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
)

type Connector struct {
	lname string
	conn  *net.UnixConn
}

func (c *Connector) Connect() (err error) {
	lname := fmt.Sprintf("/tmp/hackernel-%s.sock", uuid.New().String())
	rname := "/tmp/hackernel.sock"
	nettype := "unixgram"
	laddr := net.UnixAddr{Name: lname, Net: nettype}
	raddr := net.UnixAddr{Name: rname, Net: nettype}

	os.Remove(lname)
	conn, err := net.DialUnix(nettype, &laddr, &raddr)
	if err != nil {
		goto errout
	}
	c.lname = lname
	c.conn = conn
	return
errout:
	os.Remove(lname)
	return
}

func (c *Connector) Close() {
	c.conn.Close()
	os.Remove(c.lname)
}

func (c *Connector) Shutdown() (err error) {
	c.conn.SetReadDeadline(time.Now())
	return
}

func (c *Connector) Send(msg string) (err error) {
	n, err := c.conn.Write([]byte(msg))
	if err != nil {
		return
	}
	if n != len(msg) {
		return
	}
	return
}

func (c *Connector) Recv() (msg string, err error) {
	// 取内核能上报的最大值,(1<<12)*32 来自内核源码
	buffer := make([]byte, (1<<12)*32)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return
	}
	msg = string(buffer[0:n])
	return
}
