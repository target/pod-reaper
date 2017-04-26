package main

import "os"

func environmentVariable(key string, defValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defValue
	}
	return value
}

func loadConfiguration() configuration {
	return configuration{
		namespace:          environmentVariable("NAMESPACE", ""),
		runDuration:        environmentVariable("RUN_DURATION", "0s"),
		pollInterval:       environmentVariable("POLL_INTERVAL", "1m"),
		excludeLabelKey:    environmentVariable("EXCLUDE_LABEL_KEY", ""),
		excludeLabelValues: environmentVariable("EXCLUDE_LABEL_VALUE", ""),
	}
}

type configuration struct {
	namespace          string
	runDuration        string
	pollInterval       string
	excludeLabelKey    string
	excludeLabelValues string
}
