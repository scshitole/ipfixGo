package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bmarshall13/go-bigip-rest"
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

	//	user := flag.String("user", "admin", "username on BIG-IP")
	//	pass := flag.String("pass", "admin", "password on BIG-IP")
	tlsNoVerify := flag.Bool("skipTlsVerify", false, "Don't verifiy BigIP certificate")
	//	flag.Parse()
	//	if flag.NArg() != 1 {
	//		panic(fmt.Sprintf("Usage: f5api [--user USER] [--pass PASS] [--skipTlsVerify] <bigip>\n"))
	//	}
	//	host := flag.Arg(0)

	//	fmt.Printf("Connecting to BIG-IP %v (%v/%v)\n", host, *user, *pass)

	f5 := f5api.NewClient(Bigip_mgmt, User, Pass)
	err := f5.DoLogin()
	if err != nil {
		panic(fmt.Sprintf("Error logging in: %v", err))
	}

	// Example: list all the virtuals (compare to "show /ltm virtual")
	virtuals, err := f5.Ltm.GetVirtualList()
	if err != nil {
		panic(fmt.Sprintf("Error getting list of virtual servers: %v", err))
	}
	for _, virtual := range virtuals.Items {
		fmt.Printf("Virtual %v Destination %v\n", virtual.Name, virtual.Destination)
	}
}
