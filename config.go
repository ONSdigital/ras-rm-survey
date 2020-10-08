package main

import "github.com/spf13/viper"

func setDefaults() {
    viper.SetDefault("service_name", "ras-rm-survey")
    viper.SetDefault("app_version", "unknown")
    viper.SetDefault("dummy_health_database", "UP 100ms")
    viper.SetDefault("dummy_health_rabbitmq", "DOWN")
}