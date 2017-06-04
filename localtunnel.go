// Package localtunnel implements a client library for https://localtunnel.me
//
// localtunnel allows you to expose your localhost to the world for easy testing and sharing!
//
// Exposing a local port:
//
//    import "github.com/jweslley/localtunnel"
//
//    ...
//
//    var port := 8000
//    var tunnel := localtunnel.NewLocalTunnel(port)
//    var err := tunnel.Open()
//    if (err != nil) {
//    	fmt.Printf("your url is: %s\n", tunnel.URL())
//    }
//
//    ...
//
//    tunnel.Close()
//
//
package localtunnel

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
)

// A Client is an localtunnel client.
type Client struct {
	endPoint string
}

// NewLocalTunnel create a tunnel for a server in a given port from localhost.
func (c *Client) NewLocalTunnel(port int) *Tunnel {
	return c.NewTunnel("localhost", port)
}

// NewTunnel create a tunnel for a server in a given host and port.
func (c *Client) NewTunnel(host string, port int) *Tunnel {
	return &Tunnel{c: c, localHost: host, localPort: port}
}

// NewClient returns a client using the given end point.
func NewClient(url string) *Client {
	return &Client{endPoint: url}
}

// DefaultClient is the default Client and is used by NewLocalTunnel and NewTunnel.
var DefaultClient = NewClient("https://localtunnel.me")

// NewLocalTunnel create a tunnel for a server in a given port from localhost using the DefaultClient.
func NewLocalTunnel(port int) *Tunnel {
	return DefaultClient.NewTunnel("localhost", port)
}

// NewTunnel create a tunnel for a server in a given host and port using the DefaultClient.
func NewTunnel(host string, port int) *Tunnel {
	return DefaultClient.NewTunnel(host, port)
}

// Tunnel forwards remote requests to another server, typically to a port on localhost.
type Tunnel struct {
	c       *Client
	m       sync.Mutex
	closeCh chan struct{}

	remoteHost string
	remotePort int
	localHost  string
	localPort  int
	subdomain  string
	url        string
	maxConn    int
}

func (t *Tunnel) RemoteHost() string { return t.remoteHost }
func (t *Tunnel) RemotePort() int    { return t.remotePort }
func (t *Tunnel) LocalHost() string  { return t.localHost }
func (t *Tunnel) LocalPort() int     { return t.localPort }
func (t *Tunnel) Subdomain() string  { return t.subdomain }

// URL at which the localtunnel is exposed.
func (t *Tunnel) URL() string { return t.url }

// MaxConn is the maximum number of connections allowed.
func (t *Tunnel) MaxConn() int { return t.maxConn }

// Open setup the tunnel creating connections between the remote and local servers.
func (t *Tunnel) Open() error {
	return t.OpenAs("?new")
}

// Open setup the tunnel creating connections between the remote and local servers with a custom subdomain.
func (t *Tunnel) OpenAs(subdomain string) error {
	t.m.Lock()
	defer t.m.Unlock()

	err := t.setup(subdomain)
	if err != nil {
		return err
	}

	t.closeCh = make(chan struct{})
	t.establish()
	return nil
}

// Close closes all tunnel's connections.
func (t *Tunnel) Close() {
	t.m.Lock()
	defer t.m.Unlock()

	t.remoteHost = ""
	t.remotePort = 0
	t.maxConn = 0
	t.subdomain = ""
	t.url = ""
	close(t.closeCh)
}

// Closing is a channel which is closed when the tunnel is closed.
func (t *Tunnel) Closing() <-chan struct{} {
	return t.closeCh
}

func (t *Tunnel) setup(subdomain string) error {
	url := fmt.Sprintf(t.c.endPoint+"/%s", subdomain)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var i struct {
		ID      string `json:"id,omitempty"`
		URL     string `json:"url,omitempty"`
		Port    int    `json:"port,omitempty"`
		MaxConn int    `json:"max_conn_count,omitempty"`
	}

	d := json.NewDecoder(resp.Body)
	err = d.Decode(&i)
	if err != nil {
		return err
	}

	t.remoteHost = resp.Request.URL.Host
	t.remotePort = i.Port
	t.maxConn = i.MaxConn
	t.subdomain = i.ID
	t.url = i.URL

	return nil
}

func (t *Tunnel) establish() {
	for i := 0; i < t.MaxConn(); i++ {
		c := &conn{t: t}
		go c.open()
	}
}

type conn struct {
	t          *Tunnel
	remoteConn net.Conn
	localConn  net.Conn
}

func (c *conn) open() {
	var err error

	c.remoteConn, err = net.Dial("tcp", net.JoinHostPort(c.t.RemoteHost(), strconv.Itoa(c.t.RemotePort())))
	if err != nil {
		c.t.Close()
		return
	}

	c.localConn, err = net.Dial("tcp", net.JoinHostPort(c.t.LocalHost(), strconv.Itoa(c.t.LocalPort())))
	if err != nil {
		c.t.Close()
		return
	}

	c.pipe()
}

func (c *conn) close() {
	if c.localConn != nil {
		c.localConn.Close()
	}

	if c.remoteConn != nil {
		c.remoteConn.Close()
	}
}

func (c *conn) pipe() {
	errorCh := make(chan error)
	remoteCh := chanFromConn(c.remoteConn, errorCh)
	localCh := chanFromConn(c.localConn, errorCh)

	for {
		select {
		case b := <-remoteCh:
			c.localConn.Write(b)
		case b := <-localCh:
			c.remoteConn.Write(b)
		case <-errorCh:
			c.close()
			c.open()
			return
		case <-c.t.closeCh:
			c.close()
			return
		}
	}
}

func chanFromConn(conn net.Conn, errorCh chan error) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				errorCh <- err
				break
			}
		}
	}()

	return c
}
