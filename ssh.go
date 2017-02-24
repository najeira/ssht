package ssht

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type Address struct {
	Host string
	Port int
}

func (addr *Address) String() string {
	return fmt.Sprintf("%s:%d", addr.Host, addr.Port)
}

type Tunnel struct {
	Config *ssh.ClientConfig
	Local  Address
	Server Address
	Remote Address
}

func NewTunnel(config *ssh.ClientConfig, local, server, remote Address) *Tunnel {
	return &Tunnel{
		Config: config,
		Local:  local,
		Server: server,
		Remote: remote,
	}
}

func (tn *Tunnel) Start() error {
	localLn, err := net.Listen("tcp", tn.Local.String())
	if err != nil {
		return err
	}
	defer localLn.Close()
	//log.Println("local listen")

	for {
		if err := tn.forward(localLn); err != nil {
			return err
		}
	}
	return nil
}

func (tn *Tunnel) forward(localLn net.Listener) error {
	localConn, err := localLn.Accept()
	if err != nil {
		return err
	}
	//log.Println("local accept")

	serverConn, err := ssh.Dial("tcp", tn.Server.String(), tn.Config)
	if err != nil {
		return err
	}
	//log.Println("server dial")

	remoteConn, err := serverConn.Dial("tcp", tn.Remote.String())
	if err != nil {
		return err
	}
	//log.Println("remote dial")

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
	return nil
}

func copyConn(writer, reader net.Conn) {
	if _, err := io.Copy(writer, reader); err != nil {
		log.Println(err)
	}
}

func publicKeyAuthMethod(file string) (ssh.AuthMethod, error) {
	body, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	signers, err := ssh.ParsePrivateKey(body)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signers), nil
}

func PublicKeyAuthConfig(user string, keyFile string) (*ssh.ClientConfig, error) {
	authMethod, err := publicKeyAuthMethod(keyFile)
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{authMethod},
	}, nil
}
