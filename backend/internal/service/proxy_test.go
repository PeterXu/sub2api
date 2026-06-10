package service

import (
	"net/url"
	"testing"
)

func TestProxyURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		proxy Proxy
		want  string
	}{
		{
			name: "without auth",
			proxy: Proxy{
				Protocol: "http",
				Host:     "proxy.example.com",
				Port:     8080,
			},
			want: "http://proxy.example.com:8080",
		},
		{
			name: "with auth",
			proxy: Proxy{
				Protocol: "socks5",
				Host:     "socks.example.com",
				Port:     1080,
				Username: "user",
				Password: "pass",
			},
			want: "socks5://user:pass@socks.example.com:1080",
		},
		{
			name: "username only keeps no auth for compatibility",
			proxy: Proxy{
				Protocol: "http",
				Host:     "proxy.example.com",
				Port:     8080,
				Username: "user-only",
			},
			want: "http://proxy.example.com:8080",
		},
		{
			name: "with special characters in credentials",
			proxy: Proxy{
				Protocol: "http",
				Host:     "proxy.example.com",
				Port:     3128,
				Username: "first last@corp",
				Password: "p@ ss:#word",
			},
			want: "http://first%20last%40corp:p%40%20ss%3A%23word@proxy.example.com:3128",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.proxy.URL(); got != tc.want {
				t.Fatalf("Proxy.URL() mismatch: got=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestProxyURL_SpecialCharactersRoundTrip(t *testing.T) {
	t.Parallel()

	proxy := Proxy{
		Protocol: "http",
		Host:     "proxy.example.com",
		Port:     3128,
		Username: "first last@corp",
		Password: "p@ ss:#word",
	}

	parsed, err := url.Parse(proxy.URL())
	if err != nil {
		t.Fatalf("parse proxy URL failed: %v", err)
	}
	if got := parsed.User.Username(); got != proxy.Username {
		t.Fatalf("username mismatch after parse: got=%q want=%q", got, proxy.Username)
	}
	pass, ok := parsed.User.Password()
	if !ok {
		t.Fatal("password missing after parse")
	}
	if pass != proxy.Password {
		t.Fatalf("password mismatch after parse: got=%q want=%q", pass, proxy.Password)
	}
}

func TestInjectFromIdURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		proxy         Proxy
		injectEnabled bool
		fromId        string
		want          string
	}{
		{
			name: "inject disabled returns URL unchanged",
			proxy: Proxy{
				Protocol: "socks5", Host: "proxy.example.com", Port: 1080,
				Username: "alice", Password: "secret",
			},
			injectEnabled: false,
			fromId:        "42",
			want:          "socks5://alice:secret@proxy.example.com:1080",
		},
		{
			name: "inject enabled with username and password",
			proxy: Proxy{
				Protocol: "socks5", Host: "proxy.example.com", Port: 1080,
				Username: "alice", Password: "secret",
			},
			injectEnabled: true,
			fromId:        "42",
			want:          "socks5://alice%4042:secret@proxy.example.com:1080",
		},
		{
			name: "inject enabled with username only no password",
			proxy: Proxy{
				Protocol: "socks5", Host: "proxy.example.com", Port: 1080,
				Username: "alice",
			},
			injectEnabled: true,
			fromId:        "42",
			want:          "socks5://alice%4042@proxy.example.com:1080",
		},
		{
			name: "inject enabled but empty fromId returns URL unchanged",
			proxy: Proxy{
				Protocol: "socks5", Host: "proxy.example.com", Port: 1080,
				Username: "alice", Password: "secret",
			},
			injectEnabled: true,
			fromId:        "",
			want:          "socks5://alice:secret@proxy.example.com:1080",
		},
		{
			name: "inject enabled but no username returns URL unchanged",
			proxy: Proxy{
				Protocol: "socks5", Host: "proxy.example.com", Port: 1080,
			},
			injectEnabled: true,
			fromId:        "42",
			want:          "socks5://proxy.example.com:1080",
		},
		{
			name: "http proxy with inject enabled",
			proxy: Proxy{
				Protocol: "http", Host: "proxy.example.com", Port: 8080,
				Username: "bob", Password: "pass",
			},
			injectEnabled: true,
			fromId:        "99",
			want:          "http://bob%4099:pass@proxy.example.com:8080",
		},
		{
			name: "socks5h protocol with inject enabled",
			proxy: Proxy{
				Protocol: "socks5h", Host: "proxy.example.com", Port: 1080,
				Username: "user", Password: "pw",
			},
			injectEnabled: true,
			fromId:        "7",
			want:          "socks5h://user%407:pw@proxy.example.com:1080",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.proxy.InjectFromIdURL(tc.injectEnabled, tc.fromId); got != tc.want {
				t.Fatalf("InjectFromIdURL() mismatch: got=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestInjectFromIdURL_RoundTrip(t *testing.T) {
	t.Parallel()

	proxy := Proxy{
		Protocol: "socks5", Host: "proxy.example.com", Port: 1080,
		Username: "alice", Password: "secret",
	}

	result := proxy.InjectFromIdURL(true, "42")
	parsed, err := url.Parse(result)
	if err != nil {
		t.Fatalf("parse result URL failed: %v", err)
	}

	gotUser := parsed.User.Username()
	if gotUser != "alice@42" {
		t.Fatalf("username mismatch: got=%q want=%q", gotUser, "alice@42")
	}

	pass, ok := parsed.User.Password()
	if !ok {
		t.Fatal("password missing")
	}
	if pass != "secret" {
		t.Fatalf("password mismatch: got=%q want=%q", pass, "secret")
	}
}
