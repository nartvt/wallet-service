package util

import "strings"

func MaskEmail(str string, first, last int) string {
	idx := strings.Index(str, "@")
	if idx == -1 {
		return ""
	}
	return str[0:first] + "****" + str[idx-last:idx] + str[idx:]
}
