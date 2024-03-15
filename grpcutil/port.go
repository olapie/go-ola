package grpcutil

import (
	"slices"
	"strings"
)

var tlsPorts = []string{"443", "6443", "7743", "8443", "9443"}

func IsTLSServer(server string) bool {
	a := strings.Split(server, ":")
	return len(a) == 2 && slices.Contains(tlsPorts, a[1])
}
