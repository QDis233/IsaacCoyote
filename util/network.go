package util

import "net"

func GetLocalIP() ([]string, error) {
	var ips []string
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addresses {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil && IsPrivateIP(ipNet.IP) {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips, nil
}

func IsPrivateIP(ip net.IP) bool {
	privateBlocks := []string{
		"10.0.0.0/8",
		//"172.16.0.0/12",
		"192.168.0.0/16",
		//"169.254.0.0/16",
		//"fc00::/7",
	}

	for _, block := range privateBlocks {
		_, ipNet, _ := net.ParseCIDR(block)
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}
