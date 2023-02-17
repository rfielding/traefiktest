package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"
)

var dynamicDesiredConfig = "dynamic.yaml"
var dynamicActualConfig = "dynamic-actual.yaml"

func AsJson(o interface{}) string {
	j, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("Cannot marshal to json: %v", err))
	}
	return string(j)
}

func isUrlOk(pingUrl string) bool {
	fmt.Printf("%s\n", pingUrl)
	// We assume that the services have a GET /ping to determine rr policy
	res, err := http.Get(pingUrl)
	if err != nil {
		panic("unable to create ping url")
	}
	if err != nil {
		fmt.Printf("Ping url failed: %s %v", pingUrl, err)
		return false
	}
	if res.StatusCode >= 500 {
		return false
	}
	return true
}

// Dead loadBalancer servers need removal from the actual state
func removeDeadLoadBalancerURL(y interface{}) {
	top, topOk := y.(map[interface{}]interface{})
	if !topOk {
		panic("unable to find traefik")
	}

	topHttp, topHttpOk := top["http"].(map[interface{}]interface{})
	if !topHttpOk {
		panic("unable to find traefik.http")
	}

	topHttpServices, topHttpServicesOk := topHttp["services"].(map[interface{}]interface{})
	if !topHttpServicesOk {
		panic("unable to find traefik.http.services")
	}

	// For each http cluster....
	for k, _ := range topHttpServices {
		serviceName := k.(string)
		//fmt.Printf("# serviceName: %s:%T\n", serviceName, serviceName)

		lb, lbOk := topHttpServices[k].(map[interface{}]interface{})
		if !lbOk {
			panic(fmt.Sprintf("unable to find traefik.http.services.%s.loadBalancer", serviceName))
		}

		servers, serversOk := lb["loadBalancer"].(map[interface{}]interface{})
		if !serversOk {
			panic(fmt.Sprintf("unable to find traefik.http.services.%s.loadBalancer.servers", serviceName))
		}

		serversArray, serversArrayOk := servers["servers"].([]interface{})
		if !serversArrayOk {
			panic(fmt.Sprintf("unable to find traefik.http.services.%s.loadBalancer.servers as array", serviceName))
		}

		useServers := make([]interface{}, 0)
		for k := range serversArray {
			ua, uaOk := serversArray[k].(map[interface{}]interface{})
			if !uaOk {
				panic(fmt.Sprintf("unable to find traefik.http.services.%s.loadBalancer.servers.%d as array", serviceName, k))
			}
			u := ua["url"]
			uResolve := isUrlOk(fmt.Sprintf("%s/ping", u))
			if uResolve {
				useServers = append(useServers, serversArray[k])
			}
		}
		servers["servers"] = useServers

		newy, err := yaml.Marshal(y)
		if err != nil {
			panic(fmt.Sprintf("cannot serialize out new yaml: %v", err))
		}
		err = ioutil.WriteFile(dynamicActualConfig, newy, 0447)
		if err != nil {
			panic(fmt.Sprintf("cannot write out new yaml: %v", err))
		}
	}
}

func main() {
	prevHash := ""
	for {
		// Get the bytes of the latest dynamic desired config
		// That represents the dynamic DESIRED state
		yamlBytes, err := ioutil.ReadFile(dynamicDesiredConfig)
		if err != nil {
			fmt.Printf("Failed to load dynamic config %s: %v\n", dynamicDesiredConfig, err)
			continue
		}
		h1 := sha256.Sum256(yamlBytes)
		h2 := h1[:]
		h := hex.EncodeToString(h2)
		isDirty := prevHash != h

		var y interface{}
		var newy interface{}
		if isDirty || true {
			if isDirty || y == nil {
				fmt.Printf("# desired dynamicConfig: %s\n", h)
				// parse the yaml
				err = yaml.Unmarshal(yamlBytes, &y)
				if err != nil {
					panic(fmt.Sprintf("Unable to parse yaml: %v", err))
				}
				prevHash = h
			} else {
				fmt.Printf("# actual config\n")
			}

			removeDeadLoadBalancerURL(y)
			newy, err = yaml.Marshal(y)
			if err != nil {
				panic(fmt.Sprintf("Unable to marshal yaml: %v", err))
			}
			fmt.Printf("%s", newy)

			// make sure that every loadBalancer url exists, or is removed
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
}
