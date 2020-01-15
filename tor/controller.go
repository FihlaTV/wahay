package tor

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/wybiral/torgo"
)

const (
	// AuthTypeNotDefined if the type for non-defined Tor Control Port auth type
	AuthTypeNotDefined = ""
	// AuthTypeNone if the type for Tor Control Port `none` auth
	AuthTypeNone = "none"
	// AuthTypeCookie if the type for Tor Control Port `cookie` auth
	AuthTypeCookie = "cookie"
	// AuthTypePassword if the type for Tor Control Port `password` auth
	AuthTypePassword = "password"
)

// Control is the interface for controlling the Tor instance on this system
type Control interface {
	GetTorController() (torgoController, error)
	EnsureTorCompatibility() (bool, bool, error)
	CreateNewOnionService(destinationHost, destinationPort string, port string) (serviceID string, err error)
	DeleteOnionService(serviceID string) error
	SetInstance(i *Instance)
	Close()
}

type controller struct {
	torHost  string
	torPort  string
	password string
	authType string
	c        torgoController
	i        *Instance
	tc       func(string) (torgoController, error)
}

func (cntrl *controller) SetInstance(i *Instance) {
	cntrl.i = i
}

func (cntrl *controller) GetTorController() (torgoController, error) {
	if cntrl.c != nil {
		return cntrl.c, nil
	}

	log.Printf("Creating new Tor Control Port controller with host=%s and port=%s\n", cntrl.torHost, cntrl.torPort)
	c, err := cntrl.tc(net.JoinHostPort(cntrl.torHost, cntrl.torPort))
	if err != nil {
		return nil, err
	}

	cntrl.c = c

	return c, nil
}

// GetAuthenticationMethod checks the TOR Control Port available auth type
func GetAuthenticationMethod(tc torgoController, cntrl *controller) (string, error) {
	log.Println("Checking Tor Control Port none authentication")
	err := tc.AuthenticateNone()
	if err == nil {
		return AuthTypeNone, nil
	}
	log.Println(fmt.Sprintf("auth-none: %s", err))

	log.Println("Checking Tor Control Port cookie authentication")
	err = tc.AuthenticateCookie()
	if err == nil {
		return AuthTypeCookie, nil
	}
	log.Println(fmt.Sprintf("auth-cookie: %s", err))

	if len(cntrl.password) > 0 {
		log.Println("Checking Tor Control Port password authentication")
		err = tc.AuthenticatePassword(cntrl.password)
		if err == nil {
			return AuthTypePassword, nil
		}
		log.Println(fmt.Sprintf("auth-passw: %s", err))
	}

	addr := net.JoinHostPort(cntrl.torHost, cntrl.torPort)
	return AuthTypeNotDefined, fmt.Errorf("cannot authenticate to the TCP running on %s", addr)
}

func (cntrl *controller) EnsureTorCompatibility() (bool, bool, error) {
	tc, err := cntrl.GetTorController()
	if err != nil {
		return false, false, err
	}

	if len(cntrl.authType) == 0 {
		cntrl.authType, err = GetAuthenticationMethod(tc, cntrl)
		if err != nil {
			log.Println(err)
		}
	}

	err = Authenticate(tc, cntrl.authType, cntrl.password)
	if err != nil {
		return false, true, err
	}

	if cntrl.authType == "" {
		return false, true, errors.New("the current tor control port cannot be used")
	}

	version, err := tc.GetVersion()
	if err != nil {
		return false, true, err
	}

	diff, err := compareVersions(version, MinSupportedVersion)
	if err != nil {
		return false, true, err
	}

	if diff < 0 {
		return false, false, errors.New("version of Tor is not compatible")
	}

	return false, true, nil
}

// Authenticate make possible authentication depending of the mode
func Authenticate(tc torgoController, authType string, password string) error {
	if len(authType) == 0 {
		return errors.New("provide a specific authentication type")
	}

	switch authType {
	case AuthTypeCookie:
		return tc.AuthenticateCookie()
	case AuthTypePassword:
		return tc.AuthenticatePassword(password)
	default:
		return tc.AuthenticateNone()
	}
}

func (cntrl *controller) DeleteOnionService(serviceID string) error {
	s := strings.TrimSuffix(serviceID, ".onion")
	return cntrl.c.DeleteOnion(s)
}

func (cntrl *controller) CreateNewOnionService(destinationHost, destinationPort string,
	port string) (serviceID string, err error) {
	tc, err := cntrl.GetTorController()

	if err != nil {
		log.Println(err)
		return
	}

	err = Authenticate(tc, cntrl.authType, cntrl.password)
	if err != nil {
		return
	}

	servicePort, err := strconv.ParseUint(port, 10, 16)

	if err != nil {
		err = errors.New("invalid source port")
		return
	}

	onion := &torgo.Onion{
		Ports: map[int]string{
			int(servicePort): net.JoinHostPort(destinationHost, destinationPort),
		},
		PrivateKeyType: "NEW",
		PrivateKey:     "ED25519-V3",
	}

	err = tc.AddOnion(onion)

	if err != nil {
		return "", err
	}

	serviceID = fmt.Sprintf("%s.onion", onion.ServiceID)

	return serviceID, nil
}

func (cntrl *controller) Close() {
	if cntrl.i != nil {
		log.Println("closing our tor control port instance")
		cntrl.i.close()
	}
}

// CreateController takes the Tor information given and returns a
// controlling interface
func CreateController(torHost, torPort, password string, authType string) Control {
	f := func(v string) (torgoController, error) {
		return torgo.NewController(v)
	}

	// If password is provided, then our `authType` should
	// be `password` as the default value
	if len(authType) == 0 && len(password) > 0 {
		authType = AuthTypePassword
	}

	return &controller{
		torHost:  torHost,
		torPort:  torPort,
		password: password,
		authType: authType,
		tc:       f,
		c:        nil,
		i:        nil,
	}
}
