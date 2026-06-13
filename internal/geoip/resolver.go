package geoip

import (
	"fmt"
	"net"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

type Resolver struct {
	db            *geoip2.Reader
	defaultRegion string
}

func NewResolver(dbPath, defaultRegion string) (*Resolver, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open geoip db %s: %w", dbPath, err)
	}
	return &Resolver{
		db:            db,
		defaultRegion: strings.ToUpper(defaultRegion),
	}, nil
}

func (r *Resolver) Close() error {
	return r.db.Close()
}

func (r *Resolver) DefaultRegion() string {
	return r.defaultRegion
}

func IsLocalIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	return parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified()
}

func (r *Resolver) LookupCountry(ip string) (region string, ok bool) {
	parsed := net.ParseIP(ip)
	if parsed == nil || IsLocalIP(ip) {
		return "", false
	}

	record, err := r.db.Country(parsed)
	if err != nil {
		return "", false
	}

	code := strings.ToUpper(record.Country.IsoCode)
	if !IsValidCountryCode(code) {
		return "", false
	}
	return code, true
}
