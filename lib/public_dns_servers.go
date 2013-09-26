package carpetbomb

import "math/rand"

var PublicDnsServers = [...]string{
	"8.8.8.8:53", "8.8.4.4:53", // Google DNS
}

func GetPublicDnsServer() string {
	return PublicDnsServers[rand.Intn(len(PublicDnsServers))]
}
