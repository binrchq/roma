package utils

import "strings"

var (
	DETERMINE_IP          = "ip"
	DETERMINE_DOMAIN      = "domain"
	DETERMINE_IP_PORT     = "ipport"
	DETERMINE_DOMAIN_PORT = "domainport"
	DETERMINE_HOSTNAME    = "hostname"
)

// 确定搜索类型和资源
func DetermineSearchType(resA string) (string, string) {
	searchType := DETERMINE_HOSTNAME
	if strings.Contains(resA, ":") {
		parts := strings.Split(resA, ":")
		if IsIP(parts[0]) {
			searchType = DETERMINE_IP_PORT
		} else if IsDomain(parts[0]) {
			searchType = DETERMINE_DOMAIN_PORT
		}
	} else {
		if IsIP(resA) {
			searchType = DETERMINE_IP
		} else if IsDomain(resA) {
			searchType = DETERMINE_DOMAIN
		}
	}
	return searchType, resA
}
