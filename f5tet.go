package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/f5devcentral/go-bigip"
)

func main() {
	// Create a single reader which can be called multiple times

	var UserResponse, yourchoice, Vresponse, Eresponse, FirstSensor, SecondSensor, Uresponse, ThirdSensor string
	scanner := bufio.NewScanner(os.Stdin)
	var Bigipmgmt, User, Pass string
	fmt.Print("Enter your BIG-IP Management IP: ")
	//fmt.Print("Enter your BIG-IP Management IP: ")
	scanner.Scan()
	Bigipmgmt = scanner.Text()
	fmt.Print("Enter your Username: ")
	scanner.Scan()
	User = scanner.Text()
	fmt.Print("Enter your Password: ")
	scanner.Scan()
	Pass = scanner.Text()
	fmt.Println("Attempting to Connect...\n")
	// Establish our session to the BIG-IP
	// check irule json exists in local directory under irules/
	var i = 1
	for i > 0 {
		fmt.Print("Please make your selection 1: IPFIX Configuration\n                           2: Remove IPFIX Configuration\n                           3: Remove IPFIX iRules from Virtual Server\n                           4: Remove iRules from BIG-IP\n                           5: Exit\n\n")
		fmt.Print("Enter Your Choice : ")
		scanner.Scan()
		yourchoice = scanner.Text()
		switch yourchoice {

		case "1":

			fileTCPexists := fileTCPexists() // returns true or false
			if fileTCPexists {
				fmt.Println("TCP iRules  exists on your local machine\n")

			} else {
				fmt.Println("TCP iRules does not exists on local machine ..... getting from github\n")
				downloadTCPiruleFromGithub()

			}

			fileUDPexists := fileUDPexists() // returns true or false
			if fileUDPexists {
				fmt.Println("UDP iRules exists on your local machine\n")

			} else {
				fmt.Println("UDP iRules does not exists on local machine ..... getting from github\n")
				downloadUDPiruleFromGithub()
			}

			checkTCPonBigip := checkTCPiruleExistsOnBigip(Bigipmgmt, User, Pass)
			if checkTCPonBigip {
				fmt.Println("You have TCP iRules on BIG-IP\n")
			} else {
				//fmt.Println(" TCP Irule Does not Exists on BIG-IP")
				b, err := ioutil.ReadFile("Tetration_TCP_L4_ipfix.tcl") // just pass the file name
				if err != nil {
					fmt.Println("Not able to locate iRules on your local machine \n", err)
				}

				//fmt.Println(b) // print the content as 'bytes'

				Rule := string(b) // convert content to a 'string'

				fmt.Println("Uploading TCP iRules to BIG-IP .........\n") // print the content as a 'string'
				addTCPiruleToBigip(Bigipmgmt, User, Pass, Rule)

			}

			checkUDPonBigip := checkUDPiruleExistsOnBigip(Bigipmgmt, User, Pass)

			if checkUDPonBigip {
				fmt.Println("You have UDP iRules on BIG-IP\n")
			} else {
				//fmt.Println(" UDP Irule Does not Exists on BIG-IP")
				b, err := ioutil.ReadFile("Tetration_UDP_L4_ipfix.tcl") // just pass the file name
				if err != nil {
					fmt.Println("Not able to locate UDP iRules file on BIG-IP\n", err)
				}

				//fmt.Println(b) // print the content as 'bytes'

				Rule := string(b) // convert content to a 'string'

				fmt.Println("Uploading UDP iRules to BIG-IP .........\n") // print the content as a 'string'
				addUDPiruleToBigip(Bigipmgmt, User, Pass, Rule)

			}
			checkIpfixOnBigip := checkIpfixPoolExistsOnBigip(Bigipmgmt, User, Pass)
			if checkIpfixOnBigip == false {
				fmt.Println("IPFIX Pool Does not Exists on BIG-IP Creating .....\n")
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
				f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
				pools, err := f5.Pools()
				if err != nil {
					panic(err.Error())
				}
				for _, pool := range pools.Pools {
					//fmt.Printf("Name: %s\n", pool.Name)
					//vs.Description = "Modified Sanjay Shitole"
					if pool.Name == "TetrationIPFIXPool" {
						fmt.Printf("\033[32mName: %s\n", pool.Name)
						t, err := f5.PoolMembers("TetrationIPFIXPool")
						if err != nil {
							panic(err.Error())
						}
						for _, m := range t.PoolMembers {
							fmt.Printf("Sensors list : %s \n", m.Name)
						}

					}

				}
				fmt.Printf("\033[30m")
				checkIpfixOnBigip = true // now make it true

			}

			if checkIpfixOnBigip == true {

				fmt.Println("Above Showing you IPFIX Pool on BIG-IP \n")
				fmt.Print("Do you want to use the above shown IPFIX Pool say Y/N? ")
				scanner.Scan()
				UserResponse = scanner.Text()
				if UserResponse == "Y" || UserResponse == "y" {
					fmt.Print("Appy iRules on all Virtual Server Y/N ? : ")
					scanner.Scan()
					Vresponse = scanner.Text()
					if Vresponse == "Y" || Vresponse == "y" {
						fmt.Println("Configuring iRules on all Virtual Server ......\n")
						applyTcpIruleOnAll(Bigipmgmt, User, Pass)
						applyUdpIruleOnAll(Bigipmgmt, User, Pass)
						displayAllVirtual(Bigipmgmt, User, Pass)
					} else {
						fmt.Println("Please select which Virtual Server need iRules \n")
						applyOneByOne(Bigipmgmt, User, Pass)
						displayAllVirtual(Bigipmgmt, User, Pass)

					}
				} else {
					fmt.Print("Update Sensors  Y/N? ")
					scanner.Scan()
					Eresponse = scanner.Text()
					if Eresponse == "Y" || Eresponse == "y" {
						updateIpfixPoolMember(Bigipmgmt, User, Pass)
					} else {
						fmt.Print("Update New Sensors  Y/N? ")
						scanner.Scan()
						Uresponse = scanner.Text()
						if Uresponse == "Y" || Uresponse == "y" {
							addNewSensor(Bigipmgmt, User, Pass)
						}
					}
				}
			} // if loop end
			fmt.Print(" \n \n")
			/*fmt.Print("Do you wish to dettach the IPFIX iRules? ")
			scanner.Scan()
			dconfig = scanner.Text()
			if dconfig == "Y" || dconfig == "y" {
				DettachiRule(Bigipmgmt, User, Pass)
			}

			fmt.Print("Do you wish to Remove the IPFIX Pool Configuration? ")
			scanner.Scan()
			Vconfig = scanner.Text()
			if Vconfig == "Y" || Vconfig == "y" {
				RemovePoolConfig(Bigipmgmt, User, Pass)
			}
			DeleteiRule(Bigipmgmt, User, Pass)
			*/
		case "2":
			RemovePoolConfig(Bigipmgmt, User, Pass)
		case "3":
			DettachiRule(Bigipmgmt, User, Pass)
		case "4":
			name1 := "/Common/Tetration_TCP_L4_ipfix"
			name2 := "/Common/Tetration_UDP_L4_ipfix"
			DeleteiRule(Bigipmgmt, User, Pass, name1)
			DeleteiRule(Bigipmgmt, User, Pass, name2)
		case "5":
			os.Exit(3)
		default:
		} // switch
	} // for loop end
} // main

func checkIpfixPoolExistsOnBigip(Bigipmgmt, User, Pass string) bool {
	fmt.Println("Checking IPFIX Pool exists on BIG-IP ......\n")
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
			fmt.Printf("\033[32mName: %s\n", pool.Name)
			t, err := f5.PoolMembers("TetrationIPFIXPool")
			if err != nil {
				panic(err.Error())
			}
			for _, m := range t.PoolMembers {
				fmt.Printf("Sensors list : %s \n", m.Name)
			}
			fmt.Printf("\033[30m")
			return true

		}

	}
	return false
}

func applyTcpIruleOnAll(Bigipmgmt, User, Pass string) error {
	// Go through all the virtual servers and display them

	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	tvservers, err := f5.VirtualServers()
	if err != nil {
		panic(err.Error())
	}
	for _, vs := range tvservers.VirtualServers {
		var a = vs.Rules
		if vs.IPProtocol == "tcp" {
			vs.Rules = append(a, "/Common/Tetration_TCP_L4_ipfix") // Collect all iRules to be configured
			fmt.Printf("IPFIX TCP iRules will be applied to Virtual Server \033[32m%s\033[30m\n\n", vs.Name)
			err := f5.ModifyVirtualServer(vs.Name, &vs)
			if err != nil {
				log.Printf("\033[91m[ERROR] Unable to Apply iRule to  %s\033[30m\n\n", vs.Name)
				return err
			}

		}

	}
	return nil
}

func applyUdpIruleOnAll(Bigipmgmt, User, Pass string) error {
	// Go through all the virtual servers and display them

	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	uvservers, err := f5.VirtualServers()
	if err != nil {
		panic(err.Error())
	}

	for _, vs := range uvservers.VirtualServers {
		//fmt.Printf("%s Virtual Server type is %s and IRules on this VIP are %s\n\n ", vs.Name, vs.IPProtocol, vs.Rules)
		var a = vs.Rules
		//if  len(a) != 0 {
		if vs.IPProtocol == "udp" {
			vs.Rules = append(a, "/Common/Tetration_UDP_L4_ipfix") // Collect all iRules to be configured
			fmt.Printf("IPFIX UDP IRules will be applied to Virtual Server \033[34m%s\033[30m\n\n", vs.Name)
			err := f5.ModifyVirtualServer(vs.Name, &vs)
			if err != nil {
				log.Printf("\033[91m[ERROR] Unable to Apply iRule to  %s\033[30m\n\n", vs.Name)
				return err
			}

		}

	}
	return nil
}

func applyOneByOne(Bigipmgmt, User, Pass string) error {
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
		fmt.Printf("\033[32m%s\033[30m Virtual Server type is \033[32m%s\033[30m and IRules on this VIP are \033[32m%s\033[30m\n", vs.Name, vs.IPProtocol, vs.Rules)
		fmt.Print("Do you want to Apply iRule to this Above Virtual Server  say Y/N ? ")
		scanner.Scan()
		Uresponse = scanner.Text()
		if Uresponse == "Y" || Uresponse == "y" {
			if vs.IPProtocol == "tcp" {
				vs.Rules = append(a, "/Common/Tetration_TCP_L4_ipfix") // Collect all iRules to be configured
				fmt.Printf("IPFIX TCP IRule will be applied to Virtual Server \033[32m%s\033[30m\n\n", vs.Name)
				err := f5.ModifyVirtualServer(vs.Name, &vs)
				if err != nil {
					log.Printf("\033[91m[ERROR] Unable to Apply iRule to  %s\033[30m\n\n", vs.Name)
					return err
				}

			} else {

				if vs.IPProtocol == "udp" {
					vs.Rules = append(a, "/Common/Tetration_UDP_L4_ipfix") // Collect all iRules to be configured
					fmt.Printf("IPFIX UDP IRule will be applied to Virtual Server \033[34m%s\033[30m\n\n", vs.Name)
					err := f5.ModifyVirtualServer(vs.Name, &vs)
					if err != nil {
						log.Printf("\033[91m[ERROR] Unable to Apply iRule to  %s\033[30m\n\n", vs.Name)
						return err
					}
				} else {
					fmt.Printf("Virtual Servers is not UDP/TCP no irule applied to: %s \n", vs.Name)
				}
			}

		}
	}

	return nil
}

func displayAllVirtual(Bigipmgmt, User, Pass string) error {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	vservers, err := f5.VirtualServers()
	if err != nil {
		log.Printf("[ERROR] Unable to Read the Virtual Server  %s \n", err)
		return err
	}
	fmt.Printf("\n\n")
	fmt.Printf("Displaying all the Virtual Servers and iRules  ......\n")
	for _, vs := range vservers.VirtualServers {
		fmt.Printf("\033[32m%s\033[30m Virtual Server type is \033[32m%s\033[30m and IRules on this VIP are \033[32m%s\033[30m \n", vs.Name, vs.IPProtocol, vs.Rules)
	}
	return nil
}

func addNewSensor(Bigipmgmt, User, Pass string) error {
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
		log.Printf("[ERROR] Unable to Add a New Sensor  %s \n", err)
		return err
	}
	for _, t := range nodes.PoolMembers {
		fmt.Printf("Sensors installed are %s :\n", t.Name)
	}
	return nil

}

func updateIpfixPoolMember(Bigipmgmt, User, Pass string) error {
	var Sresponse, Dresponse string
	scanner := bufio.NewScanner(os.Stdin)
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	name := "/Common/TetrationIPFIXPool"
	members, err := f5.PoolMembers(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Read the Sensor  %s \n", err)
		return err
	}
	for _, m := range members.PoolMembers {
		fmt.Printf("Sensors installed are %s :\n", m.Name)
	}

	for _, m := range members.PoolMembers {

		fmt.Printf("Want to  change this Sensor IP %s : Y/N? : ", m.Name)
		scanner.Scan()
		Sresponse = scanner.Text()
		if Sresponse == "Y" || Sresponse == "y" {
			fmt.Print("Enter the New Sensor IP (Port not required) : ")
			scanner.Scan()
			Dresponse = scanner.Text()
			addsuccess, _ := addPoolMemebers(Bigipmgmt, User, Pass, Dresponse)
			if addsuccess {
				err := f5.DeletePoolMember(name, m.Name)
				if err != nil {
					log.Printf("[ERROR] Unable to Delete the Sensor  %s \n", err)
					return err
				}
			}

		}
	}
	t, err := f5.PoolMembers(name)
	if err != nil {
		panic(err.Error())
	}
	for _, m := range t.PoolMembers {
		fmt.Printf("Updated Sensors list : \033[32m%s\033[30m \n", m.Name)
	}
	return nil
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

func addPoolMemebers(Bigipmgmt, User, Pass, Sensor string) (bool, error) {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	member := Sensor + ":4739"
	err := f5.AddPoolMember("TetrationIPFIXPool", member)
	if err != nil {
		log.Printf("[ERROR] Unable to Add the Sensor  %s \n", member, err)
		return false, err
	}
	return true, nil
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
	fmt.Println("Checking TCP iRules  exists on your local machine\n")
	if _, err := os.Stat("Tetration_TCP_L4_ipfix.tcl"); err != nil {
		return false
	}
	return true
}

func fileUDPexists() bool {
	fmt.Println("Checking UDP iRules exists on your local machine\n")
	if _, err := os.Stat("Tetration_UDP_L4_ipfix.tcl"); err != nil {
		return false
	}
	return true
}

func checkTCPiruleExistsOnBigip(Bigipmgmt, User, Pass string) bool {
	fmt.Println("Checking TCP iRules exists on BIG-IP ......\n")
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
	fmt.Println("Checking UDP iRules exists on BIG-IP ......\n")
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
		fmt.Printf("%s Virtual Server type is %s and IRules on this VIP are %s\n\n", vs.Name, vs.IPProtocol, vs.Rules)
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
			fmt.Printf("Following iRules will be applied to Virtual Server \033[32m%s\033[30m  and iRules \033[32m%s\033[30m\n\n", vs.Name, vs.Rules)
			err := f5.ModifyVirtualServer(vs.Name, &vs)
			if err != nil {
				log.Printf("\033[91m[ERROR] Unable to Dettach iRule from  %s\033[30m\n\n", vs.Name)
				return err
			}

		}
	}
	return nil
}
func DeleteiRule(Bigipmgmt, User, Pass, name string) error {
	f5 := bigip.NewSession(Bigipmgmt, User, Pass, nil)
	fmt.Printf("Removing iRules \033[32m%s\033[30m from BIG-IP ......\n\n", name)
	err := f5.DeleteIRule(name)
	if err != nil {
		log.Printf("\033[91m[ERROR] Unable to Delete iRules, First Remove iRules %s from Virtual Server then use option 4  \033[30m\n\n", name)
		return err
	}
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
