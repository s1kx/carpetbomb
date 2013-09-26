package carpetbomb

import (
	"github.com/miekg/dns"
	"net"
)

const (
	DnsTimeout = 5e9
)

type Request struct {
	Hostname  string
	DnsServer string

	// Results
	Error       error
	IPAddresses []net.IP
}

func CreateRequest(hostname string, dnsServer string) *Request {
	ipAddresses := make([]net.IP, 0, 10)
	return &Request{hostname, dnsServer, nil, ipAddresses[:]}
}

func (r *Request) Resolve() {
	// Create DNS client
	c := new(dns.Client)
	c.ReadTimeout = DnsTimeout

	// Create DNS message
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(r.Hostname), dns.TypeA)

	// Send DNS request
	resp, _, err := c.Exchange(m, r.DnsServer)

	if err != nil {
		r.Error = err
		return
	}

	for _, answer := range resp.Answer {
		if dnsRecord, ok := answer.(*dns.A); ok {
			r.IPAddresses = append(r.IPAddresses, dnsRecord.A)
		}
	}
}
