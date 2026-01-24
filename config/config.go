package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"
)

type BackendConfig struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

type ProxyConfig struct {
	Port                 int             `json:"port"`
	Admin_port           int             `json:"admin_port"`
	Strategy             string          `json:"strategy"`
	HealthCheckFreq      time.Duration   `json:"health_check_frequency"`
	HealthCheckMethod    string          `json:"health_check_method"`
	Backend_timeout      time.Duration   `json:"backend_timeout"`
	BackendsConfig       []BackendConfig `json:"backends"`
	EnableStickySessions bool            `json:"enable_sticky_sessions"`
	StickySessionTTL     time.Duration   `json:"sticky_session_ttl"`
}

func LoadConfiguration() (p ProxyConfig, err error) {
	configuration := struct {
		Port                 int      `json:"port"`
		Admin_port           int      `json:"admin_port"`
		Strategy             string   `json:"strategy"`
		HealthCheckFreq      string   `json:"health_check_frequency"`
		HealthCheckMethod    string   `json:"health_check_method"`
		Backend_timeout      string   `json:"backend_timeout"`
		Backends             []BackendConfig `json:"backends"`
		EnableStickySessions bool     `json:"enable_sticky_sessions"`
		StickySessionTTL     string   `json:"sticky_session_ttl"`
	}{}

	jsonFile, err := os.Open("config.json")
	if err != nil {
		return ProxyConfig{}, errors.New("Error while opening the configuration file")
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &configuration)

	p.Port = configuration.Port
	p.Admin_port = configuration.Admin_port
	p.Strategy = configuration.Strategy

	p.Backend_timeout, err = time.ParseDuration(configuration.Backend_timeout)
	if err != nil {
		return ProxyConfig{}, errors.New("Error parsing timeout duration")
	}

	p.HealthCheckFreq, err = time.ParseDuration(configuration.HealthCheckFreq)
	if err != nil {
		return ProxyConfig{}, errors.New("Error parsing health check frequency ")
	}

	p.BackendsConfig = configuration.Backends

	for i := range p.BackendsConfig {
		_, err := url.Parse(p.BackendsConfig[i].URL)
		if err != nil {
			return ProxyConfig{}, errors.New("Error parsing backend url: " + p.BackendsConfig[i].URL)
		}
		if p.BackendsConfig[i].Weight == 0 {
			p.BackendsConfig[i].Weight = 1
		}
	}

	p.HealthCheckMethod = configuration.HealthCheckMethod
	p.EnableStickySessions = configuration.EnableStickySessions

	p.StickySessionTTL, err = time.ParseDuration(configuration.StickySessionTTL)
	if err != nil {
		return ProxyConfig{}, errors.New("error parsing sticky_session_ttl")
	}

	err = p.Validate()
	if err != nil {
		return ProxyConfig{}, err
	}

	return p, nil
}

func (p *ProxyConfig) Validate() error {
	if p.Port <= 0 || p.Port > 65535 {
		return errors.New("invalid port: must be between 1-65535")
	}

	if p.Admin_port <= 0 || p.Admin_port > 65535 {
		return errors.New("invalid admin_port: must be between 1-65535")
	}

	if p.Port == p.Admin_port {
		return errors.New("port and admin_port cannot be the same")
	}

	if p.Strategy != "round-robin" && p.Strategy != "least-conn" {
		fmt.Println(p.Strategy)
		return errors.New("invalid strategy: must be 'round-robin' or 'least-conn'")
	}

	if p.HealthCheckMethod != "tcp" && p.HealthCheckMethod != "http" {
		return errors.New("invalid health_check_method: must be 'tcp' or 'http'")
	}

	if p.HealthCheckFreq <= 0 {
		return errors.New("health_check_frequency must be positive")
	}

	if p.Backend_timeout <= 0 {
		return errors.New("backend_timeout must be positive")
	}

	if len(p.BackendsConfig) == 0 {
		return errors.New("at least one backend must be configured")
	}

	return nil
}