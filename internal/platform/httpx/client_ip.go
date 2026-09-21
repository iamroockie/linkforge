package httpx

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
)

type ClientIPResolver struct {
	trusted []netip.Prefix
}

func NewClientIPResolver(trustedProxyCIDRs []string) (ClientIPResolver, error) {
	resolver := ClientIPResolver{trusted: make([]netip.Prefix, 0, len(trustedProxyCIDRs))}

	for _, cidr := range trustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(cidr))
		if err != nil {
			return ClientIPResolver{}, fmt.Errorf("parse trusted proxy: %w", err)
		}

		if prefix.Addr().Is4In6() && prefix.Bits() >= 96 {
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}

		resolver.trusted = append(resolver.trusted, prefix.Masked())
	}

	return resolver, nil
}

func (r ClientIPResolver) Resolve(req *http.Request) (netip.Addr, error) {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("parse remote address: %w", err)
	}

	peer, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("parse remote IP: %w", err)
	}

	peer = peer.Unmap()
	if !r.isTrusted(peer) {
		return peer, nil
	}

	return r.resolveForwarded(req.Header.Values("X-Forwarded-For"), peer)
}

func (r ClientIPResolver) isTrusted(addr netip.Addr) bool {
	for _, prefix := range r.trusted {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func (r ClientIPResolver) resolveForwarded(values []string, peer netip.Addr) (netip.Addr, error) {
	client := peer
	for _, value := range slices.Backward(values) {
		for _, part := range slices.Backward(strings.Split(value, ",")) {
			addr, err := netip.ParseAddr(strings.TrimSpace(part))
			if err != nil {
				return netip.Addr{}, fmt.Errorf("parse forwarded IP: %w", err)
			}
			if addr.Zone() != "" {
				return netip.Addr{}, errors.New("forwarded IP must not contain a zone")
			}
			client = addr.Unmap()
			if !r.isTrusted(client) {
				return client, nil
			}
		}
	}
	return client, nil
}
