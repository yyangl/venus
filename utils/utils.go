package utils

import (
	"net"
)

func ResolveServerKey(name string, uuid string) string {
	return "/services/" + name + "/" + uuid
}

func ResolveServiceKey() string {
	return "/services"
}

func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "error"
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	panic("Unable to determine local IP address (non loopback). Exiting.")
}
