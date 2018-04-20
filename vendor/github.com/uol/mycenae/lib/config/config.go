package config

type TSDBfilterInfo struct {
	Examples    string `json:"examples"`
	Description string `json:"description"`
}

func GetAggregators() []string {
	return []string{
		"avg",
		"count",
		"min",
		"max",
		"sum",
	}
}

func GetFilters() []string {
	return []string{
		"literal_or",
		"not_literal_or",
		"wildcard",
		"regexp",
	}
}

func GetFiltersFull() map[string]TSDBfilterInfo {
	return map[string]TSDBfilterInfo{
		"literal_or": {
			Examples:    `host=iliteral_or(web01),  host=iliteral_or(web01|web02|web03)  {\"type\":\"iliteral_or\",\"tagk\":\"host\",\"filter\":\"web01|web02|web03\",\"groupBy\":false}`,
			Description: `Accepts one or more exact values and matches if the series contains any of them. Multiple values can be included and must be separated by the | (pipe) character. The filter is case insensitive and will not allow characters that TSDB does not allow at write time.`,
		},
		"not_literal_or": {
			Examples:    `host=not_literal_or(web01),  host=not_literal_or(web01|web02|web03)  {\"type\":\"not_literal_or\",\"tagk\":\"host\",\"filter\":\"web01|web02|web03\",\"groupBy\":false}`,
			Description: `Accepts one or more exact values and matches if the series does NOT contain any of them. Multiple values can be included and must be separated by the | (pipe) character. The filter is case sensitive and will not allow characters that TSDB does not allow at write time.`,
		},
		"wildcard": {
			Examples:    `host=wildcard(web*),  host=wildcard(web*.tsdb.net)  {\"type\":\"wildcard\",\"tagk\":\"host\",\"filter\":\"web*.tsdb.net\",\"groupBy\":false}`,
			Description: `Performs pre, post and in-fix glob matching of values. The globs are case sensitive and multiple wildcards can be used. The wildcard character is the * (asterisk). At least one wildcard must be present in the filter value. A wildcard by itself can be used as well to match on any value for the tag key.`,
		},
		"regexp": {
			Examples:    `host=regexp(.*)  {\"type\":\"regexp\",\"tagk\":\"host\",\"filter\":\".*\",\"groupBy\":false}`,
			Description: `Provides full, POSIX compliant regular expression using the built in Java Pattern class. Note that an expression containing curly braces {} will not parse properly in URLs. If the pattern is not a valid regular expression then an exception will be raised.`,
		},
	}
}

func GetDownsamplers() []string {
	return []string{
		"avg",
		"count",
		"min",
		"max",
		"sum",
	}
}

func GetFillers() []string {
	return []string{
		"none",
		"nan",
		"null",
		"zero",
	}
}
