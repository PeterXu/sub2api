package service

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	FallbackModeNone   = "none"
	FallbackModeProxy  = "proxy"
	FallbackModeDirect = "direct"
)

type Proxy struct {
	ID             int64
	Name           string
	Protocol       string
	Host           string
	Port           int
	Username       string
	Password       string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      *time.Time
	FallbackMode   string
	BackupProxyID  *int64
	ExpiryWarnDays int
}

func (p *Proxy) IsActive() bool {
	return p.Status == StatusActive
}

// IsExpired 报告代理是否已过期（基于 expires_at，与 status 无关）。
func (p *Proxy) IsExpired(now time.Time) bool {
	return p.ExpiresAt != nil && !p.ExpiresAt.After(now)
}

func (p *Proxy) URL() string {
	u := &url.URL{
		Scheme: p.Protocol,
		Host:   net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
	}
	if p.Username != "" && p.Password != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u.String()
}

// InjectFromIdURL returns the proxy URL with fromId injected into the username field.
// If injectEnabled is false, fromId is empty, or the proxy has no username, it falls
// back to the standard URL().
func (p *Proxy) InjectFromIdURL(injectEnabled bool, fromId string) string {
	if !injectEnabled || fromId == "" || p.Username == "" {
		return p.URL()
	}
	u := &url.URL{
		Scheme: p.Protocol,
		Host:   net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
	}
	injectedUser := p.Username + "@" + fromId
	if p.Password != "" {
		u.User = url.UserPassword(injectedUser, p.Password)
	} else {
		u.User = url.User(injectedUser)
	}
	return u.String()
}

func (p *Proxy) NewURL(c *gin.Context, account *Account) string {
	injectEnabled := account.IsInjectUserIdInProxyEnabled()
	fromId := c.GetHeader("X-Proxy-User-Id")
	return p.InjectFromIdURL(injectEnabled, fromId)
}

type ProxyWithAccountCount struct {
	Proxy
	AccountCount   int64
	LatencyMs      *int64
	LatencyStatus  string
	LatencyMessage string
	IPAddress      string
	Country        string
	CountryCode    string
	Region         string
	City           string
	QualityStatus  string
	QualityScore   *int
	QualityGrade   string
	QualitySummary string
	QualityChecked *int64
}

type ProxyAccountSummary struct {
	ID       int64
	Name     string
	Platform string
	Type     string
	Notes    *string
}
