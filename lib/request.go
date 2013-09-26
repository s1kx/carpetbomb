package carpetbomb

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
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
	dnsClient := new(dns.Client)

	// Create DNS message
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(r.Hostname), dns.TypeA)

	// Send DNS request
	response, _, err := dnsClient.Exchange(msg, r.DnsServer)

	if err != nil {
		r.Error = err
		return
	}

	if len(response.Answer) == 0 {
		r.Error = errors.New(fmt.Sprintf("Invalid hostname: %s", r.Hostname))
		return
	}

	for _, answer := range response.Answer {
		if dnsRecord, ok := answer.(*dns.A); ok {
			// Get Record
			ipAddress := dnsRecord.A

			r.IPAddresses = append(r.IPAddresses, ipAddress)
		}
	}
}
