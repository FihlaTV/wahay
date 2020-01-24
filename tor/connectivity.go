package tor

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"autonomia.digital/tonio/app/config"
	"github.com/wybiral/torgo"
	"golang.org/x/net/proxy"
)

const UsrBinPath = "/usr/bin/tor"
const UsrLocalBinPath = "/usr/local/bin/tor"

// Connectivity is used to check whether Tor can connect in different ways
type Connectivity interface {
	Check() (total error, partial error)
}

type connectivity struct {
	host        string
	checkBinary bool
	routePort   int
	controlPort int
	password    string
}

// NewDefaultChecker will test whether the default ports can
// be reached and are appropriate for our use
func NewDefaultChecker() Connectivity {
	// This checks everything, including binaries against the default ports
	return NewChecker(true, *config.TorHost, *config.TorRoutePort, *config.TorPort, *config.TorControlPassword)
}

// NewChecker can check connectivity on custom ports, and optionally
// avoid checking for binary compatibility
func NewChecker(checkBinary bool, host string, routePort, controlPort int, password string) Connectivity {
	return &connectivity{
		host:        host,
		checkBinary: checkBinary,
		routePort:   routePort,
		controlPort: controlPort,
		password:    password,
	}
}

func (c *connectivity) checkTorControlPortExists() bool {
	_, err := torgo.NewController(net.JoinHostPort(c.host, strconv.Itoa(c.controlPort)))
	return err == nil
}

func withNewTorgoController(where string, a authenticationMethod) authenticationMethod {
	return func(torgoController) error {
		tc, err := torgo.NewController(where)
		if err != nil {
			return err
		}
		return a(tc)
	}
}

func (c *connectivity) checkTorControlAuth() bool {
	where := net.JoinHostPort(c.host, strconv.Itoa(c.controlPort))

	authCallback := authenticateAny(
		withNewTorgoController(where, authenticateNone),
		withNewTorgoController(where, authenticateCookie),
		withNewTorgoController(where, authenticatePassword(c.password)))

	return authCallback(nil) == nil
}

type checkTorResult struct {
	IsTor bool
	IP    string
}

func (c *connectivity) checkConnectionOverTor() bool {
	proxyURL, err := url.Parse("socks5://" + net.JoinHostPort(c.host, strconv.Itoa(c.routePort)))
	if err != nil {
		return false
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return false
	}

	t := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{Transport: t}

	resp, err := client.Get("https://check.torproject.org/api/ip")
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	var v checkTorResult
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return false
	}

	return v.IsTor
}

func (c *connectivity) Check() (total error, partial error) {
	b := GetTorBinary(nil)

	err := b.Check()
	if err != nil {
		return err, nil
	}

	if !c.checkTorControlPortExists() {
		return nil, errors.New("no Tor Control Port found")
	}

	if !c.checkTorControlAuth() {
		return nil, errors.New("no Tor Control Port valid authentication")
	}

	err = b.Check()
	if err != nil {
		return err, nil
	}

	if !c.checkConnectionOverTor() {
		return errors.New("not connection over Tor allowed"), nil
	}

	return nil, nil
}
