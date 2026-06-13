package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

func main() {
	dbPath := flag.String("db", "data/GeoLite2-Country.mmdb", "path to GeoLite2-Country.mmdb")
	ip := flag.String("ip", "", "IP address to lookup")
	flag.Parse()

	if strings.TrimSpace(*ip) == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/geoip-lookup --ip 1.2.3.4 [--db path/to/GeoLite2-Country.mmdb]")
		os.Exit(1)
	}

	parsed := net.ParseIP(strings.TrimSpace(*ip))
	if parsed == nil {
		log.Fatalf("invalid ip: %q", *ip)
	}

	reader, err := geoip2.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer reader.Close()

	record, err := reader.Country(parsed)
	if err != nil {
		log.Fatalf("lookup: %v", err)
	}

	fmt.Printf("ip=%s\n", parsed.String())
	fmt.Printf("country_code=%s\n", strings.ToUpper(record.Country.IsoCode))
	fmt.Printf("country_name=%s\n", record.Country.Names["en"])
}
