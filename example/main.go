package main

import (
	"github.com/xmidt-org/carousel"
	"fmt"
	"regexp"
	"strconv"
)

var hostnameRegex = regexp.MustCompile(`^[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?(?:\.[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?)*\.?$`)
var prefixRegex = regexp.MustCompile(`(.*?)\.`)

var Check carousel.HostValidator = CheckHost

func CheckHost(fqdn string) bool {
	matches := hostnameRegex.FindStringSubmatch(fqdn)
	if matches == nil {
		// not a valid hostname
		fmt.Printf("%s is not a valid hostname.\n", fqdn)
		return false
	}
	prefexMatches := prefixRegex.FindStringSubmatch(fqdn)
	if prefexMatches == nil {
		// no Matches
		fmt.Printf("no matches for regex %s.\n", prefixRegex.String())
		return false
	}
	c := prefexMatches[1][len(prefexMatches[1])-1:]
	val, err := strconv.Atoi(c)
	if err != nil {
		fmt.Printf("%s is not a number. \n", c)
		return false
	}
	if val%2 != 0 {
		fmt.Printf("%d is not even. \n", val)
		return false
	}
	return true
}
