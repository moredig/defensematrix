package deception

import (
	"fmt"
	"net/http"
)

// BaitProfile defines the fake identity a Boat broadcasts
type BaitProfile struct {
	ServerBanner string
	FakeHeaders  map[string]string
	OpenPorts    []int
	DefaultCreds map[string]string // username -> password
}

// Profiles returns a set of tempting fake server identities
func Profiles() map[string]*BaitProfile {
	return map[string]*BaitProfile{
		"apache-legacy": {
			ServerBanner: "Apache/2.2.8 (Ubuntu) PHP/5.2.4-2ubuntu5",
			FakeHeaders: map[string]string{
				"Server":           "Apache/2.2.8 (Ubuntu) PHP/5.2.4-2ubuntu5",
				"X-Powered-By":     "PHP/5.2.4-2ubuntu5",
				"X-AspNet-Version": "2.0.50727",
			},
			OpenPorts:    []int{22, 80, 443, 3306, 5432},
			DefaultCreds: map[string]string{"admin": "admin", "root": "root", "user": "password"},
		},
		"iis-legacy": {
			ServerBanner: "Microsoft-IIS/6.0",
			FakeHeaders: map[string]string{
				"Server":           "Microsoft-IIS/6.0",
				"X-Powered-By":     "ASP.NET",
				"X-AspNet-Version": "1.1.4322",
			},
			OpenPorts:    []int{80, 443, 1433, 3389},
			DefaultCreds: map[string]string{"administrator": "admin", "sa": "sa"},
		},
		"nginx-exposed": {
			ServerBanner: "nginx/1.10.3",
			FakeHeaders: map[string]string{
				"Server": "nginx/1.10.3",
			},
			OpenPorts:    []int{22, 80, 8080, 9200},
			DefaultCreds: map[string]string{"admin": "1234", "elastic": ""},
		},
	}
}

// InjectHeaders applies fake headers to an HTTP response
func InjectHeaders(w http.ResponseWriter, profile *BaitProfile) {
	for key, val := range profile.FakeHeaders {
		w.Header().Set(key, val)
	}
	fmt.Printf("[WAMAI] Bait headers injected — posing as: %s\n", profile.ServerBanner)
}
