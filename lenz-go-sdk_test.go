// main_test.go

package lenzsdk

import (
	"testing"
)

func TestIPAddress(t *testing.T) {
	// valid IP test
	// resp := IPValidator("192.168.0.0")
	if !IPValidator("192.168.0.0") {
		t.Fatalf("error in regex with valid IP")
	}

	// unvalid IP addresses for test
	unvalidIPs := []string{
		"192.168.0.1 some thingh else",
		"192.168.0.256",
		"192.168.256.255",
	}

	for _, item := range unvalidIPs {
		if IPValidator(item) {
			t.Fatalf("error in regex with unvalid IP: " + item)
		}
	}
}
