package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Create a single reader which can be called multiple times
	reader := bufio.NewReader(os.Stdin)
	// Prompt and read
	fmt.Print("Enter BIG-IP Management IP: ")
	Bigip_mgmt, _ := reader.ReadString('\n')
	fmt.Print("Enter username for BIG-IP: ")
	User, _ := reader.ReadString('\n')
	fmt.Print("Enter password  for BIG-IP: ")
	Pass, _ := reader.ReadString('\n')
	fmt.Print("Enter IPFIX Pool  for BIG-IP: ")
	Ipfix_pool, _ := reader.ReadString('\n')

	// Trim whitespace and print
	fmt.Printf("Bigip Management is: \"%s\", Bigip username : \"%s\", Bigip pass : \"%s\", Bigip IPFIX Pool \"%s\"\n",
		strings.TrimSpace(Bigip_mgmt), strings.TrimSpace(User), strings.TrimSpace(Pass), strings.TrimSpace(Ipfix_pool))
	fmt.Printf(Bigip_mgmt)
	create := createipfx()
	fmt.Println(create)
}

func createipfx() string {
	type Payload struct {
		Name    string `json:"name"`
		Monitor string `json:"monitor"`
		Members []struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"members"`
	}
	data := Payload{
		// fill struct
		Name:    "FirstIPFIXPool",
		Monitor: "gateway_icmp ",
		Members: []struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		}{
			{
				Name:    "Ipfix_pool:4739",
				Address: "Ipfix_pool",
			},
		},
	}
	fmt.Printf("value of data %+v", data)
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", os.ExpandEnv("https://10.192.74.73/mgmt/tm/ltm/pool"), body)
	fmt.Printf("value of req is  ", req)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth("admin", "admin")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	fmt.Printf("value of response is %+v ", resp)

	/* if err != nil {
	   return err
	     // handle err
	    }a*/
	defer resp.Body.Close()
	return "Created ipfix pool"

}
