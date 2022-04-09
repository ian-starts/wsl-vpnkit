package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
)

func TestHandler(t *testing.T) {
	type args struct {
		name     string
		question dns.Question
		expected []dns.RR
		code     int
	}
	tests := []args{
		{
			name: "NS resolved",
			question: dns.Question{
				Qtype: dns.TypeNS,
				Name:  "example.com.",
			},
			expected: []dns.RR{&dns.NS{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeNS,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Ns: "a.iana-servers.net.",
			}},
			code: dns.RcodeSuccess,
		},
		{
			name: "NS error",
			question: dns.Question{
				Qtype: dns.TypeNS,
				Name:  "test.com.",
			},
			expected: nil,
			code:     dns.RcodeNameError,
		},
		{
			name: "A IPv4 resolved",
			question: dns.Question{
				Qtype: dns.TypeA,
				Name:  "example.com.",
			},
			expected: []dns.RR{&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: net.ParseIP("123.123.123.123").To4(),
			}},
			code: dns.RcodeSuccess,
		},
		{
			name: "A IPv6 resolved",
			question: dns.Question{
				Qtype: dns.TypeA,
				Name:  "ipv6.com.",
			},
			expected: nil,
			code:     dns.RcodeSuccess,
		},
		{
			name: "A error",
			question: dns.Question{
				Qtype: dns.TypeA,
				Name:  "test.com.",
			},
			expected: nil,
			code:     dns.RcodeNameError,
		},
		{
			name: "AAAA ignored",
			question: dns.Question{
				Qtype: dns.TypeAAAA,
				Name:  "example.com.",
			},
			expected: nil,
			code:     dns.RcodeSuccess,
		},
		{
			name: "Zone A resolved",
			question: dns.Question{
				Qtype: dns.TypeA,
				Name:  "host.internal.",
			},
			expected: []dns.RR{&dns.A{
				Hdr: dns.RR_Header{
					Name:   "host.internal.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: net.ParseIP("192.168.1.1").To4(),
			}},
			code: dns.RcodeSuccess,
		},
		{
			name: "Zone DefaultIP A resolved",
			question: dns.Question{
				Qtype: dns.TypeA,
				Name:  "unknown.internal.",
			},
			expected: []dns.RR{&dns.A{
				Hdr: dns.RR_Header{
					Name:   "unknown.internal.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: net.ParseIP("192.168.1.100").To4(),
			}},
			code: dns.RcodeSuccess,
		},
		{
			name: "Zone A unresolved",
			question: dns.Question{
				Qtype: dns.TypeA,
				Name:  "host.error.",
			},
			expected: nil,
			code:     dns.RcodeNameError,
		},
		{
			name: "Zone NS ignored",
			question: dns.Question{
				Qtype: dns.TypeNS,
				Name:  "host.internal.",
			},
			expected: nil,
			code:     dns.RcodeSuccess,
		},
		{
			name: "Writer error",
			question: dns.Question{
				Qtype: dns.TypeNS,
				Name:  "write.error.",
			},
			expected: nil,
			code:     dns.RcodeSuccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &mockResponseWriter{}
			h := NewHandler([]types.Zone{
				{
					Name:      "internal.",
					DefaultIP: net.ParseIP("192.168.1.100"),
					Records: []types.Record{
						{Name: "host", IP: net.ParseIP("192.168.1.1")},
					},
				},
				{
					Name: "error.",
				},
			}, &mockResolver{})
			h.ServeDNS(w, &dns.Msg{Question: []dns.Question{tt.question}})
			expected := dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response:           true,
					RecursionAvailable: true,
					Rcode:              tt.code,
				},
				Question: []dns.Question{tt.question},
				Answer:   tt.expected,
			}
			if !cmp.Equal(expected, *w.Msg) {
				t.Error(cmp.Diff(expected, *w.Msg))
			}
		})
	}
}

type mockResolver struct{}

func (r *mockResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if host == "example.com." {
		return []net.IPAddr{
			{IP: net.ParseIP("123.123.123.123")},
		}, nil
	}
	if host == "ipv6.com." {
		return []net.IPAddr{
			{IP: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
		}, nil
	}
	return nil, errors.New(fmt.Sprintf("unexpected host %s", host))
}

func (r *mockResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	if name == "example.com." {
		return []*net.NS{
			{Host: "a.iana-servers.net."},
		}, nil
	}
	return nil, errors.New(fmt.Sprintf("unexpected name %s", name))
}

type mockResponseWriter struct {
	Msg *dns.Msg
}

func (w *mockResponseWriter) LocalAddr() net.Addr {
	return nil
}

func (w *mockResponseWriter) RemoteAddr() net.Addr {
	return nil
}

func (w *mockResponseWriter) WriteMsg(m *dns.Msg) error {
	w.Msg = m
	if m.Question[0].Name == "write.error." {
		return errors.New("write error")
	}
	return nil
}

func (w *mockResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (w *mockResponseWriter) Close() error {
	return errors.New("not implemented")
}

func (w *mockResponseWriter) TsigStatus() error {
	return errors.New("not implemented")
}

func (w *mockResponseWriter) TsigTimersOnly(bool) {}

func (w *mockResponseWriter) Hijack() {}
