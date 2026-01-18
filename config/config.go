package config

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"time"
)

type ProxyConfig struct {

	Port int `json:"port"`
	Admin_port int `json:"admin_port"`
	Strategy string `json:"strategy"` // e.g., "round-robin" or "least-conn"
	HealthCheckFreq time.Duration  `json:"health_check_frequency"`
	HealthCheckMethod string `json:"health_check_method"`
	Backend_timeout time.Duration `json:"backend_timeout"`
	BackendsUrls []*url.URL `json:"backends"`
    
	

}


func LoadConfiguration() (p ProxyConfig,err error){


	configuration := struct{

		Port int `json:"port"`
		Admin_port int `json:"admin_port"`
		Strategy string `json:"strategy"` // e.g., "round-robin" or "least-conn"
		HealthCheckFreq string  `json:"health_check_frequency"`
		HealthCheckMethod string `json:"health_check_method"`
		Backend_timeout string `json:"backend_timeout"`
		BackendsUrls [] string `json:"backends"`

		}{

		}
		jsonFile,err := os.Open("config.json")
		if err != nil{
			return ProxyConfig{}, errors.New("Error while opening the configuration file")
		}
		defer jsonFile.Close()
		byteValue,_ := io.ReadAll(jsonFile)
		json.Unmarshal(byteValue,&configuration)
		p.Port = configuration.Port
		p.Admin_port =configuration.Admin_port
		p.Backend_timeout,err = time.ParseDuration(configuration.Backend_timeout)
		if err != nil{
			return ProxyConfig{}, errors.New("Error parsing timeout duration")
		}
		p.HealthCheckFreq,err = time.ParseDuration(configuration.HealthCheckFreq)
		if err != nil{
			return ProxyConfig{},errors.New("Error parsing health check frequency ")
		}
		
		for i := 0; i<len(configuration.BackendsUrls);i++{
			parsed,err :=  url.Parse(configuration.BackendsUrls[i])
			if err != nil{
				return ProxyConfig{}, errors.New("Error parsing backends urls")
			}
			p.BackendsUrls = append(p.BackendsUrls,parsed)
		}
		p.HealthCheckMethod = configuration.HealthCheckMethod
		err = p.Validate()
		if err != nil{
			return ProxyConfig{}, err
		}
		return p,nil

	}

	
func (p *ProxyConfig) Validate() error {
    // Validate ports
    if p.Port <= 0 || p.Port > 65535 {
        return errors.New("invalid port: must be between 1-65535")
    }
    
    if p.Admin_port <= 0 || p.Admin_port > 65535 {
        return errors.New("invalid admin_port: must be between 1-65535")
    }
    
    if p.Port == p.Admin_port {
        return errors.New("port and admin_port cannot be the same")
    }
    
    // Validate strategy
    if p.Strategy != "round-robin" && p.Strategy != "least-conn" {
        return errors.New("invalid strategy: must be 'round-robin' or 'least-conn'")
    }
    
    // Validate health check method
    if p.HealthCheckMethod != "tcp" && p.HealthCheckMethod != "http" {
        return errors.New("invalid health_check_method: must be 'tcp' or 'http'")
    }
    
    // Validate durations
    if p.HealthCheckFreq <= 0 {
        return errors.New("health_check_frequency must be positive")
    }
    
    if p.Backend_timeout <= 0 {
        return errors.New("backend_timeout must be positive")
    }
    
    // Validate backends
    if len(p.BackendsUrls) == 0 {
        return errors.New("at least one backend must be configured")
    }
    
    return nil
}