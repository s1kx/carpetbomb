package carpetbomb

import "math/rand"

var PublicDnsServers = [...]string{
	"8.8.8.8:53", "8.8.4.4:53", // Google DNS
	"208.67.222.222:53", "208.67.220.220:53", // OpenDNS
	"209.244.0.3:53", "209.244.0.4:53", // Level3
	"156.154.70.1:53", "156.154.71.1:53", // DNS Advantage
	"4.2.2.1:53", "4.2.2.2:53", // Verizon (Level3)
	"74.122.198.48:53", "72.14.183.109:53", // OpenNIC
	"69.164.196.21:53", "208.115.243.38:53", //
}

func GetRandomPublicDnsServer() string {
	return PublicDnsServers[rand.Intn(len(PublicDnsServers))]
}
