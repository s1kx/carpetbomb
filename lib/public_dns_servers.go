package carpetbomb

import "math/rand"

var PublicDnsServers = [...]string{
	"8.8.8.8:53", "8.8.4.4:53", // Google DNS
	"208.67.222.222:53", "208.67.220.220:53", // OpenDNS
}

func GetPublicDnsServer() string {
	return PublicDnsServers[rand.Intn(len(PublicDnsServers))]
}
