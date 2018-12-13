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

	var UserResponse, Vresponse, Eresponse, FirstSensor, SecondSensor, Uresponse, Vconfig, ThirdSensor string
	scanner := bufio.NewScanner(os.Stdin)
	var Bigipmgmt, User, Pass string
	fmt.Print("Enter your bigipmgmt: ")
	scanner.Scan()
	Bigipmgmt = scanner.Text()
	fmt.Print("Enter your user: ")
	scanner.Scan()
	User = scanner.Text()
	fmt.Print("Enter your pass: ")
	scanner.Scan()
	Pass = scanner.Text()
	fmt.Println("Attempting to connect...\n")
	// Establish our session to the BIG-IP
	//f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)

	// check irule json exists in local directory under irules/

	fileTCPexists := fileTCPexists() // returns true or false
	if fileTCPexists {
		fmt.Println("Checking TCP iRule  exists on your local machine\n")

	} else {
		fmt.Println("TCP irule does not exists on local machine ..... getting from github\n")
		downloadTCPiruleFromGithub()

	}

	fileUDPexists := fileUDPexists() // returns true or false
	if fileUDPexists {
		fmt.Println("Checking UDP iRule exists on your local machine\n")

	} else {
		fmt.Println("UDP irule does not exists on local machine ..... getting from github\n")
		downloadUDPiruleFromGithub()
	}

	checkTCPonBigip := checkTCPiruleExistsOnBigip(Bigipmgmt, User, Pass)
	if checkTCPonBigip {
		fmt.Println("You have TCP irule on BIG-IP\n")
	} else {
		//fmt.Println(" TCP Irule Does not Exists on BIG-IP")
		b, err := ioutil.ReadFile("irules/Tetration_TCP_L4_ipfix.tcl") // just pass the file name
		if err != nil {
			fmt.Println("Not able to locate irule on your local machine \n", err)
		}

		//fmt.Println(b) // print the content as 'bytes'

		Rule := string(b) // convert content to a 'string'

		fmt.Println("Uploading TCP Irule to BIG-IP .........\n") // print the content as a 'string'
		addTCPiruleToBigip(Bigipmgmt, User, Pass, Rule)

	}

	checkUDPonBigip := checkUDPiruleExistsOnBigip(Bigipmgmt, User, Pass)

	if checkUDPonBigip {
		fmt.Println("You have UDP irule on BIG-IP\n")
	} else {
		//fmt.Println(" UDP Irule Does not Exists on BIG-IP")
		b, err := ioutil.ReadFile("irules/Tetration_UDP_L4_ipfix.tcl") // just pass the file name
		if err != nil {
			fmt.Println("Not able to locate UDP irule file on BIG-IP\n", err)
		}

		//fmt.Println(b) // print the content as 'bytes'

		Rule := string(b) // convert content to a 'string'

		fmt.Println("Uploading UDP Irule to BIG-IP .........\n") // print the content as a 'string'
		addUDPiruleToBigip(Bigipmgmt, User, Pass, Rule)

	}
	checkIpfixOnBigip := checkIpfixPoolExistsOnBigip(Bigipmgmt, User, Pass)
	if checkIpfixOnBigip == false {
		fmt.Println("IPFIX Pool Does not Exists on BIGIP Creating .....\n")
		fmt.Print("Enter first IPFIX Sensor : ")
		scanner.Scan()
		FirstSensor = scanner.Text()
		fmt.Print("Enter Second IPFIX Sensor : ")
		scanner.Scan()
		SecondSensor = scanner.Text()
		fmt.Print("Enter Third IPFIX Sensor : ")
		scanner.Scan()
		ThirdSensor = scanner.Text()
		createNewIPfixPool(Bigipmgmt, User, Pass)
		addPoolMemebers(Bigipmgmt, User, Pass, FirstSensor)
		addPoolMemebers(Bigipmgmt, User, Pass, SecondSensor)
		addPoolMemebers(Bigipmgmt, User, Pass, ThirdSensor)
		fmt.Println("Created .... IPFIX Pool and Members added \n\n")
		createIPFIXLog(Bigipmgmt, User, Pass)
		createPublisher(Bigipmgmt, User, Pass)
		checkIpfixOnBigip = true // now make it true

	}

	if checkIpfixOnBigip == true {

		fmt.Println("IPFIX Pool Exists on BIG-IP already\n")
		fmt.Print("Do you want to use Existing IPFIX Pool say Y/N? ")
		scanner.Scan()
		UserResponse = scanner.Text()
		if UserResponse == "Y" {
			fmt.Print("Appy iRule on all Virtual Server Y/N ? : ")
			scanner.Scan()
			Vresponse = scanner.Text()
			if Vresponse == "Y" {
				fmt.Println("Configuring iRule on all Virtual Server ......\n")
				applyIruleOnAll(Bigipmgmt, User, Pass)
			} else {
				fmt.Println("Please select which Virtual Server need iRule \n")
				applyOneByOne(Bigipmgmt, User, Pass)
			}
		} else {
			fmt.Print("Update Sensors  Y/N? ")
			scanner.Scan()
			Eresponse = scanner.Text()
			if Eresponse == "Y" {
				updateIpfixPoolMember(Bigipmgmt, User, Pass)
			} else {
				fmt.Print("Update New Sensors  Y/N? ")
				scanner.Scan()
				Uresponse = scanner.Text()
				if Uresponse == "Y" {
					addNewSensor(Bigipmgmt, User, Pass)
				}
			}
		}
	} // if loop end
	fmt.Print("#### DO YOU WISH TO REMOVE CONFIGURATION RECENTLY DONE ###### Y/N ? ")
	scanner.Scan()
	Vconfig = scanner.Text()
	if Vconfig == "Y" {
		RemovePoolConfig(Bigipmgmt, User, Pass)
		DettachiRule(Bigipmgmt, User, Pass)
		DeleteiRule(Bigipmgmt, User, Pass)
	}
}

func checkIpfixPoolExistsOnBigip(Bigipmgmt, User, Pass string) bool {
	fmt.Println("Checking IPFIX Pool exists on bigip ......\n")
	// Iterate over all the Pools, and display their names.
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)

	pools, err := f5.Pools()
	if err != nil {
		panic(err.Error())
	}

	for _, pool := range pools.Pools {
		//fmt.Printf("Name: %s\n", pool.Name)
		//vs.Description = "Modified Sanjay Shitole"
		if pool.Name == "TetrationIPFIXPool" {
			fmt.Printf("Name: %s\n", pool.Name)
			t, err := f5.PoolMembers("TetrationIPFIXPool")
			if err != nil {
				panic(err.Error())
			}
			for _, m := range t.PoolMembers {
				fmt.Printf("Sensors list : %s \n", m.Name)
			}
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
		fmt.Printf("%s Virtual Server type is %s and IRules on this VIP are %s\n ", vs.Name, vs.IPProtocol, vs.Rules)
		fmt.Print("Do you want to Apply iRule to this Above Virtual Server  say Y/N ? ")
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
	fmt.Print("Enter the New Sensor IP (Port not required): ")
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

		fmt.Printf("Want to  change this Sensor IP %s : Y/N? : ", m.Name)
		scanner.Scan()
		Sresponse = scanner.Text()
		if Sresponse == "Y" {
			err := f5.DeletePoolMember(name, m.Name)
			if err != nil {
				panic(err.Error())
			}
			fmt.Print("Enter the New Sensor IP (Port not required) : ")
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
	if err != nil {
		return err
	}
	monitor := "gateway_icmp"
	t := f5.AddMonitorToPool(monitor, name)
	if t != nil {
		fmt.Printf("Error in applying monitor %s to pool %s", monitor, name)
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

func createIPFIXLog(Bigipmgmt, User, Pass string) error {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	fmt.Println("Creating IPFIX Log Destination ......")
	err := f5.CreateLogIPFIX("TetrationIPFIXLog", "", "TetrationIPFIXPool", "ipfix", "", 5, 30, "udp")
	if err != nil {
		return err
	}
	return nil
}

func createPublisher(Bigipmgmt, User, Pass string) error {
	var p bigip.LogPublisher
	p.Name = "ipfix-pub-1"
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	p.Dests = make([]bigip.Destinations, 0, 1)
	var r bigip.Destinations
	r.Name = "TetrationIPFIXLog"
	r.Partition = "Common"
	p.Dests = append(p.Dests, r)

	fmt.Println("Creating Log Publisher  ......")

	err := f5.CreateLogPublisher(&p)
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
	fmt.Println("Checking TCP irule exists on bigip ......\n")
	// Iterate over all the iRules, and display their names.
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)

	irules, err := f5.IRules()
	if err != nil {
		panic(err.Error())
	}

	for _, irule := range irules.IRules {
		//fmt.Printf("Name: %s\n", irule.Name)
		//vs.Description = "Modified Sanjay Shitole"
		if irule.Name == "Tetration_TCP_L4_ipfix" {
			return true
		}

	}
	return false
}

func checkUDPiruleExistsOnBigip(Bigipmgmt, User, Pass string) bool {
	fmt.Println("Checking UDP irule exists on bigip ......\n")
	// Iterate over all the iRules, and display their names.
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)

	irules, err := f5.IRules()
	if err != nil {
		panic(err.Error())
	}

	for _, irule := range irules.IRules {
		//	fmt.Printf("Name:  %s\n", irule.Name)
		//vs.Description = "Modified Sanjay Shitole"

		if irule.Name == "Tetration_UDP_L4_ipfix" {
			return true
		}

	}
	return false
}
func RemovePoolConfig(Bigipmgmt, User, Pass string) error {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	fmt.Println("Removing Publisher Configuration ........")
	name := "ipfix-pub-1"
	err := f5.DeleteLogPublisher(name)
	if err != nil {
		return err
	}
	fmt.Println("Removing IPFIX log Configuration ........")

	t := f5.DeleteLogIPFIX("TetrationIPFIXLog")
	if t != nil {
		return t
	}
	fmt.Println("Removing IPFIX Pool Configuration ........")

	p := f5.DeletePool("TetrationIPFIXPool")
	if p != nil {
		return p
	}
	return nil
}

func DettachiRule(Bigipmgmt, User, Pass string) error {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	vservers, err := f5.VirtualServers()
	if err != nil {
		panic(err.Error())
	}

	for _, vs := range vservers.VirtualServers {
		fmt.Printf("%s Virtual Server type is %s and IRules on this VIP are %s\n\n ", vs.Name, vs.IPProtocol, vs.Rules)
		var a = vs.Rules
		if (vs.IPProtocol == "tcp") || (vs.IPProtocol == "udp") {
			i := 0
			for i < len(a) { //looping from 0 to the length of the array
				if a[i] == "/Common/Tetration_TCP_L4_ipfix" {
					vs.Rules[i] = ""
				}
				if a[i] == "/Common/Tetration_UDP_L4_ipfix" {
					vs.Rules[i] = ""
				}
				i++
			}
			vs.Rules = a // Collect all iRules to be configured
			fmt.Printf("Following IRule will be applied to Virtual Server %s and iRules %s\n\n ", vs.Name, vs.Rules)
			err := f5.ModifyVirtualServer(vs.Name, &vs)
			if err != nil {
				return err
			}

		}
	}
	return nil
}
func DeleteiRule(Bigipmgmt, User, Pass string) error {
	return nil
}
func downloadTCPiruleFromGithub() bool {
	fmt.Println("Downloading from github ........")
	fileUrl := "https://raw.githubusercontent.com/f5devcentral/f5-tetration/master/irules/Tetration_TCP_L4_ipfix.tcl"
	err := DownloadFile("Tetration_TCP_L4_ipfix.tcl", fileUrl)
	if err != nil {
		panic(err)
	}
	return true
}

func downloadUDPiruleFromGithub() bool {
	fmt.Println("Downloading from github ........")
	fileUrl := "https://raw.githubusercontent.com/f5devcentral/f5-tetration/master/irules/Tetration_UDP_L4_ipfix.tcl"
	err := DownloadFile("Tetration_UDP_L4_ipfix.tcl", fileUrl)
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
	if err != nil {
		return false
	}

	return true
}

func addUDPiruleToBigip(Bigipmgmt, User, Pass, Rule string) bool {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	name := "/Common/Tetration_UDP_L4_ipfix"
	err := f5.CreateIRule(name, Rule)
	if err != nil {
		return false
	}

	return true
}
