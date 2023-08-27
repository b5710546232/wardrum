package utils

import (
	"fmt"
	"regexp"
	"strings"
)

func MatchesWildcard(eventName, pattern string) bool {
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	match, err := regexp.MatchString("^"+regexPattern+"$", eventName)
	if err != nil {
		// Handle regex error if needed
		fmt.Println(err)
		return false
	}
	return match
}
