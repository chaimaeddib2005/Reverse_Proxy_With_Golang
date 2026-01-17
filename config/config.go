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
	Backend_timeout time.Duration `json:"backend_timeout"`
	BackendsUrls []*url.URL `json:"backends"`
    
	

}


func LoadConfiguration() (p ProxyConfig,err error){


	configuration := struct{

		Port int `json:"port"`
		Admin_port int `json:"admin_port"`
		Strategy string `json:"strategy"` // e.g., "round-robin" or "least-conn"
		HealthCheckFreq string  `json:"health_check_frequency"`
		Backend_timeout string `json:"backend_timeout"`
		BackendsUrls [] string `json:"backends"`

		}{

		}
		jsonFile,err := os.Open("config.json")
		if err != nil{
			return p, errors.New("Error while opening the configuration file")
		}
		defer jsonFile.Close()
		byteValue,_ := io.ReadAll(jsonFile)
		json.Unmarshal(byteValue,&configuration)
		p.Port = configuration.Port
		p.Admin_port =configuration.Admin_port
		p.Backend_timeout,err = time.ParseDuration(configuration.Backend_timeout)
		if err != nil{
			return p, errors.New("Error parsing timeout duration")
		}
		p.HealthCheckFreq,err = time.ParseDuration(configuration.HealthCheckFreq)
		if err != nil{
			return p,errors.New("Error parsing health check frequency ")
		}
		
		for i := 0; i<len(configuration.BackendsUrls);i++{
			parsed,err :=  url.Parse(configuration.BackendsUrls[i])
			if err != nil{
				return p, errors.New("Error parsing backends urls")
			}
			p.BackendsUrls = append(p.BackendsUrls,parsed)
		}

		return p,nil

	}