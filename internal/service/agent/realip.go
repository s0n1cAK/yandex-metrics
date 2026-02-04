package agent

import (
	"net"
)

func localIPForRemote(remoteHostPort string) (string, error) {
	conn, err := net.Dial("udp", remoteHostPort)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || udpAddr.IP == nil {
		return "", nil
	}
	return udpAddr.IP.String(), nil
}
