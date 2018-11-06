package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/f5devcentral/go-bigip"
)

func main() {
	// Create a single reader which can be called multiple times
	var UserResponse, Vresponse, Eresponse, FirstSensor, SecondSensor, Uresponse, ThirdSensor string
	scanner := bufio.NewScanner(os.Stdin)
	/* var Bigipmgmt, User, Pass string
	fmt.Print("Enter your bigipmgmt: ")
	scanner.Scan()
	Bigipmgmt = scanner.Text()
	fmt.Print("Enter your user: ")
	scanner.Scan()
	User = scanner.Text()
	fmt.Print("Enter your pass: ")
	scanner.Scan()
	Pass = scanner.Text() */

	// check irule json exists in local directory under irules/

	fileTCPexists := fileTCPexists() // returns true or false
	if fileTCPexists {
		fmt.Println("TCP iRule  exists on your local machine")

	} else {
		fmt.Println("TCP irule does not exists on local machine ..... getting from github")
		downloadTCPiruleFromGithub()

	}

	fileUDPexists := fileUDPexists() // returns true or false
	if fileUDPexists {
		fmt.Println("UDP iRule exists on your local machine")

	} else {
		fmt.Println("UDP irule does not exists on local machine ..... getting from github")
		downloadUDPiruleFromGithub()
	}

	Bigipmgmt := "10.192.74.73"
	User := "admin"
	Pass := "admin"
	fmt.Println("Attempting to connect...")
	// Establish our session to the BIG-IP
	//f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	checkTCPonBigip := checkTCPiruleExistsOnBigip(Bigipmgmt, User, Pass)
	if checkTCPonBigip {
		fmt.Println("TCP irule Exists on BIG-IP")
	} else {
		//fmt.Println(" TCP Irule Does not Exists on BIG-IP")
		b, err := ioutil.ReadFile("irules/Tetration_TCP_L4_ipfix.tcl") // just pass the file name
		if err != nil {
			fmt.Print("Not able to locate irule on your machine ", err)
		}

		//fmt.Println(b) // print the content as 'bytes'

		Rule := string(b) // convert content to a 'string'

		fmt.Println("Uploading TCP Irule to BIG-IP .........") // print the content as a 'string'
		addTCPiruleToBigip(Bigipmgmt, User, Pass, Rule)

	}
	checkUDPonBigip := checkUDPiruleExistsOnBigip(Bigipmgmt, User, Pass)

	if checkUDPonBigip {
		fmt.Println("UDP irule Exists on BIG-IP")
	} else {
		//fmt.Println(" UDP Irule Does not Exists on BIG-IP")
		b, err := ioutil.ReadFile("irules/Tetration_UDP_L4_ipfix.tcl") // just pass the file name
		if err != nil {
			fmt.Print("Not able to locate UDP irule file on your local machine ", err)
		}

		//fmt.Println(b) // print the content as 'bytes'

		Rule := string(b) // convert content to a 'string'

		fmt.Println("Uploading UDP Irule to BIG-IP .........") // print the content as a 'string'
		addUDPiruleToBigip(Bigipmgmt, User, Pass, Rule)

	}

	checkIpfixOnBigip := checkIpfixPoolExistsOnBigip(Bigipmgmt, User, Pass)

	if checkIpfixOnBigip {
		fmt.Println("IPFIX Pool Exists on BIG-IP already")
		fmt.Println("Do you want to use Existing IPFIX Pool say Y/N ?")
		scanner.Scan()
		UserResponse = scanner.Text()
		if UserResponse == "Y" {
			fmt.Println("Appy iRule on all Virtual Server Y/N ?")
			scanner.Scan()
			Vresponse = scanner.Text()
			if Vresponse == "Y" {
				fmt.Println("Configuring iRule on all Virtual Server ......")
				applyIruleOnAll(Bigipmgmt, User, Pass)
			} else {
				fmt.Println("Please select which Virtual Server need iRule \n")
				applyOneByOne(Bigipmgmt, User, Pass)
			}
		} else {
			fmt.Println("Update Sensors  Y/N?")
			scanner.Scan()
			Eresponse = scanner.Text()
			if Eresponse == "Y" {
				updateIpfixPoolMember(Bigipmgmt, User, Pass)
			} else {
				fmt.Println("Update New Sensors  Y/N?")
				scanner.Scan()
				Uresponse = scanner.Text()
				if Uresponse == "Y" {
					addNewSensor(Bigipmgmt, User, Pass)
				}
			}

		}

	} else {
		fmt.Println("IPFIX Pool Does not Exists on BIGIP Creating .....")
		fmt.Println("Enter first IPFIX Sensor")
		scanner.Scan()
		FirstSensor = scanner.Text()
		fmt.Println("Enter Second IPFIX Sensor")
		scanner.Scan()
		SecondSensor = scanner.Text()
		fmt.Println("Enter Third IPFIX Sensor")
		scanner.Scan()
		ThirdSensor = scanner.Text()
		createNewIPfixPool(Bigipmgmt, User, Pass)
		addPoolMemebers(Bigipmgmt, User, Pass, FirstSensor)
		addPoolMemebers(Bigipmgmt, User, Pass, SecondSensor)
		addPoolMemebers(Bigipmgmt, User, Pass, ThirdSensor)
		fmt.Println("Created .... IPFIX Pool and added Members \n\n")
	}

}

func checkIpfixPoolExistsOnBigip(Bigipmgmt, User, Pass string) bool {
	fmt.Println("Checking IPFIX Pool exists on bigip ......")
	// Iterate over all the Pools, and display their names.
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)

	pools, err := f5.Pools()
	if err != nil {
		panic(err.Error())
	}

	for _, pool := range pools.Pools {
		fmt.Printf("Name: %s\n", pool.Name)
		//vs.Description = "Modified Sanjay Shitole"
		if pool.Name == "TetrationIPFIXPool" {
			return true
		}

	}
	return false
}

func applyIruleOnAll(Bigipmgmt, User, Pass string) {
	// Go through all the virtual servers and display them

	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	vservers, err := f5.VirtualServers()
	if err != nil {
		panic(err.Error())
	}

	for _, vs := range vservers.VirtualServers {
		fmt.Printf("%s Virtual Server type is %s and IRules on this VIP are %s\n\n ", vs.Name, vs.IPProtocol, vs.Rules)
		var a = vs.Rules
		//if  len(a) != 0 {
		if vs.IPProtocol == "tcp" {
			vs.Rules = append(a, "/Common/Tetration_TCP_L4_ipfix") // Collect all iRules to be configured
			fmt.Printf("IPFIX TCP IRule will be applied to Virtual Server %s\n\n ", vs.Name)
			err := f5.ModifyVirtualServer(vs.Name, &vs)
			if err != nil {
				return
			}

		} else {

			if vs.IPProtocol == "udp" {
				vs.Rules = append(a, "/Common/Tetration_UDP_L4_ipfix") // Collect all iRules to be configured
				fmt.Printf("IPFIX UDP IRule will be applied to Virtual Server %s\n\n ", vs.Name)
				//	vs.Rules = []string{"/Common/Tetration_UDP_L4_ipfix"}
				f5.ModifyVirtualServer(vs.Name, &vs)
			} else {
				fmt.Printf("Virtual Servers is not UDP/TCP no irule applied to: %s \n", vs.Name)
			}
		}
		//   } // inner if  loop

	} // outer for loop

}

func applyOneByOne(Bigipmgmt, User, Pass string) {
	var Uresponse string
	scanner := bufio.NewScanner(os.Stdin)
	// Go through all the virtual servers and display them
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	vservers, err := f5.VirtualServers()
	if err != nil {
		panic(err.Error())
	}
	for _, vs := range vservers.VirtualServers {
		var a = vs.Rules
		fmt.Printf("%s Virtual Server type is %s and IRules on this VIP are %s\n\n ", vs.Name, vs.IPProtocol, vs.Rules)
		fmt.Println("Do you want to Apply iRule to this Above Virtual Server  say Y/N ?")
		scanner.Scan()
		Uresponse = scanner.Text()
		if Uresponse == "Y" {
			if vs.IPProtocol == "tcp" {
				vs.Rules = append(a, "/Common/Tetration_TCP_L4_ipfix") // Collect all iRules to be configured
				fmt.Printf("IPFIX TCP IRule will be applied to Virtual Server %s\n\n ", vs.Name)
				err := f5.ModifyVirtualServer(vs.Name, &vs)
				if err != nil {
					return
				}

			} else {

				if vs.IPProtocol == "udp" {
					vs.Rules = append(a, "/Common/Tetration_UDP_L4_ipfix") // Collect all iRules to be configured
					fmt.Printf("IPFIX UDP IRule will be applied to Virtual Server %s\n\n ", vs.Name)
					//	vs.Rules = []string{"/Common/Tetration_UDP_L4_ipfix"}
					f5.ModifyVirtualServer(vs.Name, &vs)
				} else {
					fmt.Printf("Virtual Servers is not UDP/TCP no irule applied to: %s \n", vs.Name)
				}
			}

		}
	}

	//return nil
}

func addNewSensor(Bigipmgmt, User, Pass string) {
	var Nresponse string
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	name := "/Common/TetrationIPFIXPool"
	members, err := f5.PoolMembers(name)
	if err != nil {
		panic(err.Error())
	}
	for _, m := range members.PoolMembers {
		fmt.Printf("You have following Sensors installed already  %s :\n", m.Name)
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter the New Sensor IP (Port not required) \n")
	scanner.Scan()
	Nresponse = scanner.Text()
	addPoolMemebers(Bigipmgmt, User, Pass, Nresponse)
	nodes, err := f5.PoolMembers(name)
	if err != nil {
		panic(err.Error())
	}
	for _, t := range nodes.PoolMembers {
		fmt.Printf("Sensors installed are %s :\n", t.Name)
	}

}

func updateIpfixPoolMember(Bigipmgmt, User, Pass string) {
	var Sresponse, Dresponse string
	scanner := bufio.NewScanner(os.Stdin)
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	name := "/Common/TetrationIPFIXPool"
	members, err := f5.PoolMembers(name)
	if err != nil {
		panic(err.Error())
	}
	for _, m := range members.PoolMembers {
		fmt.Printf("Sensors installed are %s :\n", m.Name)
	}

	for _, m := range members.PoolMembers {

		fmt.Printf("Want to  change this Sensor IP %s : Y/N? \n", m.Name)
		scanner.Scan()
		Sresponse = scanner.Text()
		if Sresponse == "Y" {
			err := f5.DeletePoolMember(name, m.Name)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("Enter the New Sensor IP (Port not required) \n")
			scanner.Scan()
			Dresponse = scanner.Text()
			addPoolMemebers(Bigipmgmt, User, Pass, Dresponse)
		}
	}
	t, err := f5.PoolMembers(name)
	if err != nil {
		panic(err.Error())
	}
	for _, m := range t.PoolMembers {
		fmt.Printf("Updated Sensors list : %s \n", m.Name)
	}
}

func createNewIPfixPool(Bigipmgmt, User, Pass string) error {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	name := "/Common/TetrationIPFIXPool"
	err := f5.CreatePool(name)

	fmt.Println(err)
	if err != nil {
		return err
	}
	return nil
}

func addPoolMemebers(Bigipmgmt, User, Pass, Sensor string) error {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	member := Sensor + ":4739"
	err := f5.AddPoolMember("TetrationIPFIXPool", member)
	if err != nil {
		return err
	}
	return nil
}

func fileTCPexists() bool {
	if _, err := os.Stat("irules/Tetration_TCP_L4_ipfix.tcl"); err != nil {
		return false
	}
	return true
}

func fileUDPexists() bool {
	if _, err := os.Stat("irules/Tetration_UDP_L4_ipfix.tcl"); err != nil {
		return false
	}
	return true
}

func checkTCPiruleExistsOnBigip(Bigipmgmt, User, Pass string) bool {
	fmt.Println("Checking TCP irule exists on bigip ......")
	// Iterate over all the iRules, and display their names.
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)

	irules, err := f5.IRules()
	if err != nil {
		panic(err.Error())
	}

	for _, irule := range irules.IRules {
		fmt.Printf("Name: %s\n", irule.Name)
		//vs.Description = "Modified Sanjay Shitole"
		if irule.Name == "Tetration_TCP_L4_ipfix" {
			//fmt.Println(" Tetration_TCP_L4_ipfix does not  Exists")
			return true
		}

	}
	return false
}

func checkUDPiruleExistsOnBigip(Bigipmgmt, User, Pass string) bool {
	fmt.Println("Checking UDP irule exists on bigip ......")
	// Iterate over all the iRules, and display their names.
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)

	irules, err := f5.IRules()
	if err != nil {
		panic(err.Error())
	}

	for _, irule := range irules.IRules {
		fmt.Printf("Name:  %s\n", irule.Name)
		//vs.Description = "Modified Sanjay Shitole"

		if irule.Name == "Tetration_UDP_L4_ipfix" {
			//fmt.Println(" Tetration_UDP_L4_ipfix Does not Exists")
			return true
		}

	}
	return false
}

func downloadTCPiruleFromGithub() bool {
	fmt.Println("Downloading from github ........")
	fileUrl := "https://raw.githubusercontent.com/f5devcentral/f5-tetration/master/irules/Tetration_TCP_L4_ipfix.tcl"
	err := DownloadFile("irules/Tetration_TCP_L4_ipfix.tcl", fileUrl)
	if err != nil {
		panic(err)
	}
	return true
}

func downloadUDPiruleFromGithub() bool {
	fmt.Println("Downloading from github ........")
	fileUrl := "https://raw.githubusercontent.com/f5devcentral/f5-tetration/master/irules/Tetration_UDP_L4_ipfix.tcl"
	err := DownloadFile("irules/Tetration_UDP_L4_ipfix.tcl", fileUrl)
	if err != nil {
		panic(err)
	}
	return true
}

//DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func addTCPiruleToBigip(Bigipmgmt, User, Pass, Rule string) bool {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	name := "/Common/Tetration_TCP_L4_ipfix"
	err := f5.CreateIRule(name, Rule)
	fmt.Println(err)
	if err != nil {
		return false
	}

	return true
}

func addUDPiruleToBigip(Bigipmgmt, User, Pass, Rule string) bool {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	name := "/Common/Tetration_UDP_L4_ipfix"
	err := f5.CreateIRule(name, Rule)
	fmt.Println(err)
	if err != nil {
		return false
	}

	return true
}