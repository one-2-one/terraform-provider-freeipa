package utils

import "strings"

// IsMembermanagerGroupDecodeError reports whether the given error originates from
// go-freeipa failing to decode the MembermanagerGroup field returned by IPA.
func IsMembermanagerGroupDecodeError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "MembermanagerGroup")
}
