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
			Name:  "Shell Injection",
			Regex: regexp.MustCompile(`(?i)(;|\||\|\||&&)\s*(cat|ls|pwd|whoami|wget|curl|bash|sh)`),
			Level: ThreatHigh,
		},
	}
}
