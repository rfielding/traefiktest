package main

import (
	"fmt"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
	"crypto/sha256"
	"bytes"
)

var dynamicConfig = "dynamic.yaml"
var dynamicActualConfig = "dynamic-actual.yaml"

func AsJson(o interface{}) string {
	j, err := json.MarshalIndent(o, "","  ")
	if err != nil {
		panic(fmt.Sprintf("Cannot marshal to json: %v", err))
	}
	return string(j)
}

func main() {
	prevHash := []byte("")
	for {
		// Get the bytes of the latest dynamic desired config
		// That represents the dynamic DESIRED state
		yamlBytes, err := ioutil.ReadFile(dynamicConfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to load dynamic config %s: %v", dynamicConfig, err))
		}
		h := sha256.Sum256(yamlBytes)

		if bytes.Compare(prevHash,h[:]) != 0 {
			fmt.Printf("changed %s\n", dynamicConfig)
			// parse the yaml
			var y interface{}
			err = yaml.Unmarshal(yamlBytes, &y)
			if err != nil {
				panic(fmt.Sprintf("Unable to parse yaml: %v", err))
			}
			prevHash = h[:]

			// make sure that every loadBalancer url exists, or is removed
			time.Sleep(time.Duration(1) * time.Second)
		} 
	}
}
