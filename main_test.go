package main

import "testing"

func TestEnsurePort(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"hostname without port", "example.com", "example.com:22"},
		{"hostname with port", "example.com:2222", "example.com:2222"},
		{"localhost without port", "localhost", "localhost:22"},
		{"localhost with port", "localhost:22", "localhost:22"},
		{"ipv4 without port", "192.168.1.1", "192.168.1.1:22"},
		{"ipv4 with port", "192.168.1.1:2222", "192.168.1.1:2222"},
		{"ipv6 without port", "::1", "[::1]:22"},
		{"ipv6 without port full", "2001:db8::1", "[2001:db8::1]:22"},
		{"ipv6 with port", "[::1]:2222", "[::1]:2222"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensurePort(tt.host)
			if got != tt.want {
				t.Errorf("ensurePort(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}
