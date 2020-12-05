package lenzsdk

import (
	"os"
	"regexp"
)

// IPValidator checks the input string with regex based on real IP Adress Version 4 like 192.168.0.0
func IPValidator(ipAddress string) bool {
	match, _ := regexp.MatchString(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`, ipAddress)
	return match
}

// HostName return the host name
func HostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
