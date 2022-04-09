package dns

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
}

type dnsHandler struct {
	zones    []types.Zone
	resolver resolver
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true
	h.addAnswers(m)
	if err := w.WriteMsg(m); err != nil {
		log.Error(err)
	}
}

func (h *dnsHandler) addAnswers(m *dns.Msg) {
	for _, q := range m.Question {
		for _, zone := range h.zones {
			zoneSuffix := fmt.Sprintf(".%s", zone.Name)
			if strings.HasSuffix(q.Name, zoneSuffix) {
				h.handleZone(zone, zoneSuffix, q, m)
				return
			}
		}

		switch q.Qtype {
		case dns.TypeNS:
			h.handleTypeNS(context.TODO(), q, m)
		case dns.TypeA:
			h.handleTypeA(context.TODO(), q, m)
		}
	}
}

func (h *dnsHandler) handleZone(zone types.Zone, zoneSuffix string, q dns.Question, m *dns.Msg) {
	if q.Qtype != dns.TypeA {
		return
	}
	for _, record := range zone.Records {
		withoutZone := strings.TrimSuffix(q.Name, zoneSuffix)
		if (record.Name != "" && record.Name == withoutZone) ||
			(record.Regexp != nil && record.Regexp.MatchString(withoutZone)) {
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: record.IP,
			})
			return
		}
	}
	if !zone.DefaultIP.Equal(net.IP("")) {
		m.Answer = append(m.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    0,
			},
			A: zone.DefaultIP,
		})
		return
	}
	m.Rcode = dns.RcodeNameError
}

func (h *dnsHandler) handleTypeNS(ctx context.Context, q dns.Question, m *dns.Msg) {
	records, err := h.resolver.LookupNS(ctx, q.Name)
	if err != nil {
		m.Rcode = dns.RcodeNameError
		return
	}
	for _, ns := range records {
		m.Answer = append(m.Answer, &dns.NS{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeNS,
				Class:  dns.ClassINET,
				Ttl:    0,
			},
			Ns: ns.Host,
		})
	}
}

func (h *dnsHandler) handleTypeA(ctx context.Context, q dns.Question, m *dns.Msg) {
	ips, err := h.resolver.LookupIPAddr(ctx, q.Name)
	if err != nil {
		m.Rcode = dns.RcodeNameError
		return
	}
	for _, ip := range ips {
		if len(ip.IP.To4()) != net.IPv4len {
			continue
		}
		m.Answer = append(m.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    0,
			},
			A: ip.IP.To4(),
		})
	}
}

func NewHandler(zones []types.Zone, resolver resolver) dns.Handler {
	mux := dns.NewServeMux()
	handler := &dnsHandler{
		zones:    zones,
		resolver: resolver,
	}
	mux.HandleFunc(".", handler.ServeDNS)
	return mux
}
