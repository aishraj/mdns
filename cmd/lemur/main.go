package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"encoding/json"
	"io/ioutil"

	"github.com/aishraj/mdns"
)

type serviceConfig struct {
	ServiceTag string   `json:"serviceTag"`
	Domain     string   `json:"domain"`
	Hostname   string   `json:"hostName"`
	Port       int      `json:"port"`
	Info       string   `json:"info"`
	Txt        []string `json:"txt"`
}

type config struct {
	Services []serviceConfig `json:"services"`
}

func main() {
	configPath := flag.String("config", "~/.mdns_services.json", "Configuration YAML file containing the services mapping")
	flag.Parse()
	log.Println("Using the config file at: ", *configPath)
	conf, err := parseConfig(*configPath)
	if err != nil {
		log.Panic("Unable to open the config file.", err)
	}

	if len(conf.Services) < 1 {
		log.Panic("The configuration should have at least one service setup.")
	}

	host, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	var services []mdns.Zone
	for _, v := range conf.Services {
		service, err := mdns.NewMDNSService(host, v.ServiceTag, v.Domain, v.Hostname, v.Port, nil, v.Txt)
		if err != nil {
			log.Fatal(err)
		}
		services = append(services, service)
	}

	server, err := mdns.NewServer(&mdns.Config{Zones: services})
	defer server.Shutdown()
	wait()

}

func wait() {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
}

func parseConfig(path string) (config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return config{}, err
	}
	var cfg config
	json.Unmarshal(file, &cfg)
	return cfg, nil
}
