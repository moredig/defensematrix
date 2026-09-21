package detection

import "regexp"

// LoadSignatures returns all known malicious patterns
func LoadSignatures() []*Signature {
	return []*Signature{
		{
			Name:  "SQL Injection",
			Regex: regexp.MustCompile(`(?i)(union\s+select|drop\s+table|insert\s+into|select\s+\*|or\s+1=1)`),
			Level: ThreatHigh,
		},
		{
			Name:  "Directory Traversal",
			Regex: regexp.MustCompile(`(\.\./|\.\.\\|%2e%2e%2f|%2e%2e/)`),
			Level: ThreatHigh,
		},
		{
			Name:  "XSS Attempt",
			Regex: regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=|alert\()`),
			Level: ThreatMedium,
		},
		{
			Name:  "Brute Force",
			Regex: regexp.MustCompile(`(?i)(failed\s+password|invalid\s+user|authentication\s+failure)`),
			Level: ThreatMedium,
		},
		{
			Name:  "Port Scan",
			Regex: regexp.MustCompile(`(?i)(nmap|masscan|zmap|nikto|dirbuster)`),
			Level: ThreatLow,
		},
		{
			Name:  "Shell Injection",
			Regex: regexp.MustCompile(`(?i)(;|\||\|\||&&)\s*(cat|ls|pwd|whoami|wget|curl|bash|sh)`),
			Level: ThreatHigh,
		},
		{
			Name:  "Credential Stuffing",
			Regex: regexp.MustCompile(`(?i)(admin|root|password|passwd|credentials)`),
			Level: ThreatLow,
		},
	}
}
