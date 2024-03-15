package grpcutil

import (
	"slices"
	"strings"
)

var tlsPorts = []string{"443", "7743", "8443", "9443"}

func IsTLSServer(server string) bool {
	strs := strings.Split(server, ":")
	return len(strs) == 2 && slices.Contains(tlsPorts, strs[1])
}
