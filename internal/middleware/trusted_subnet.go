package middleware

import (
	"log"
	"net"
	"net/http"
)

func TrustedSubnetMiddleware(next http.Handler, trustedSubnet string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if trustedSubnet == "" {
			next.ServeHTTP(w, r)
			return
		}

		realIP := r.Header.Get("X-Real-IP")
		if realIP == "" {
			log.Printf("Missing X-Real-IP header")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		_, trustedNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			log.Printf("Invalid trusted subnet CIDR: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ip := net.ParseIP(realIP)
		if ip == nil {
			log.Printf("Invalid IP address in X-Real-IP header: %s", realIP)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !trustedNet.Contains(ip) {
			log.Printf("IP %s is not in trusted subnet %s", realIP, trustedSubnet)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
