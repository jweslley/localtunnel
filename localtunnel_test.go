package localtunnel

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"
)

var ltRegexp = regexp.MustCompile("^https:\\/\\/.*\\.localtunnel.me$")

func TestDefaultClient(t *testing.T) {
	if DefaultClient == nil {
		t.Fatal("DefaultClient can not be null")
	}

	if DefaultClient.endPoint != "https://localtunnel.me" {
		t.Fatalf("Unexpected default remote host: %s", DefaultClient.endPoint)
	}
}

func TestSetupLocalTunnel(t *testing.T) {
	content := "Hello from local server!"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, content)
	}))

	localPort := getServerPort(t, s)

	tunnel := NewLocalTunnel(localPort)

	checkTunnelIsNotConnected(t, tunnel, localPort)

	err := tunnel.Open()
	if err != nil {
		t.Fatalf("Cannot open tunnel: %s", err)
	}

	checkTunnelIsConnected(t, tunnel, localPort)

	for i := 0; i <= tunnel.MaxConn()*2; i++ {
		response, err := readFromURL(tunnel.URL())
		if err != nil {
			t.Fatalf("Cannot connect through the tunnel: %s", err)
		}
		if response != content {
			t.Fatalf("Unexpected response. Expected: '%s'. Actual: '%s'", content, response)
		}
	}

	tunnel.Close()
	<-tunnel.Closing()
	checkTunnelIsNotConnected(t, tunnel, localPort)
}

func checkTunnelIsNotConnected(t *testing.T, tunnel *Tunnel, localPort int) {
	if tunnel.RemoteHost() != "" {
		t.Fatalf("Remote host should be empty: %s", tunnel.RemoteHost())
	}

	if tunnel.RemotePort() != 0 {
		t.Fatalf("Remote port should be zero. Actual: %d", tunnel.RemotePort())
	}

	if tunnel.LocalHost() != "localhost" {
		t.Fatalf("Local host should be localhost. Actual: %s", tunnel.LocalHost())
	}

	if tunnel.LocalPort() != localPort {
		t.Fatalf("Unexpected local port. Expected: %d, Actual: %d", localPort, tunnel.LocalPort())
	}

	if tunnel.Subdomain() != "" {
		t.Fatalf("Subdomain should be empty. Actual: %s", tunnel.Subdomain())
	}

	if tunnel.URL() != "" {
		t.Fatalf("URL should be empty. Actual: %s", tunnel.URL())
	}

	if tunnel.MaxConn() != 0 {
		t.Fatalf("Max connections should be zero. Actual: %d", tunnel.MaxConn())
	}
}

func checkTunnelIsConnected(t *testing.T, tunnel *Tunnel, localPort int) {
	if tunnel.RemoteHost() == "" {
		t.Fatalf("Remote host should not be empty: %s", tunnel.RemoteHost())
	}

	if tunnel.RemotePort() <= 0 {
		t.Fatalf("Remote port should be greater than zero. Actual: %d", tunnel.RemotePort())
	}

	if tunnel.LocalHost() != "localhost" {
		t.Fatalf("Local host should be localhost. Actual: %s", tunnel.LocalHost())
	}

	if tunnel.LocalPort() != localPort {
		t.Fatalf("Unexpected local port. Expected: %d, Actual: %d", localPort, tunnel.LocalPort())
	}

	if tunnel.Subdomain() == "" {
		t.Fatalf("Subdomain should not be empty. Actual: %s", tunnel.Subdomain())
	}

	if tunnel.URL() == "" {
		t.Fatalf("URL should not be empty. Actual: %s", tunnel.URL())
	}

	if !ltRegexp.MatchString(tunnel.URL()) {
		t.Fatalf("URL should match '%s'. Actual: %s", ltRegexp, tunnel.URL())
	}

	if tunnel.MaxConn() <= 0 {
		t.Fatalf("Max connections should be greater than zero. Actual: %d", tunnel.MaxConn())
	}
}

func readFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func getServerPort(t *testing.T, s *httptest.Server) int {
	url, e := url.Parse(s.URL)
	if e != nil {
		t.Fatal(e)
	}

	_, portStr, e := net.SplitHostPort(url.Host)
	if e != nil {
		t.Fatal(e)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(e)
	}

	if port <= 0 {
		t.Fatalf("Local port should be greater than zero. Actual: %d", port)
	}
	return port
}
