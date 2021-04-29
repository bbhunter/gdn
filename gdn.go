package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/projectdiscovery/httpx/common/fileutil"

	tld "github.com/jpillora/go-tld"
)

// Timeout to check for a connection - in seconds
var timout time.Duration = 2

func getips(args []string) []string {
	var inp_ips []string
	if fileutil.HasStdin() && len(args) == 1 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if len(text) != 0 {
				inp_ips = append(inp_ips, text)
			}
		}

	} else if len(args) == 2 {
		filename := args[1]
		if fileutil.FileExists(filename) {
			ips := fileutil.LoadFile(filename)
			inp_ips = append(inp_ips, ips...)
		}
	}
	return inp_ips

}

func gethostname(ip_port string) string {

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	var domainname string
	conn, err := net.DialTimeout("tcp", ip_port, timout*time.Second)
	if err == nil {
		tlsconn := tls.Client(conn, conf)
		handshake := tlsconn.Handshake()
		if handshake == nil {
			state := tlsconn.ConnectionState()
			hostname := state.PeerCertificates[0].Subject.CommonName
			hostname = "https://" + hostname
			u, errr := tld.Parse(hostname)
			if errr == nil {
				if u.Subdomain == "*" || u.Subdomain == "" {
					domainname = u.Domain + "." + u.TLD
				} else {
					domainname = u.Subdomain + "." + u.Domain + "." + u.TLD
				}
			}
			tlsconn.Close()
		}
		conn.Close()
	}

	return domainname
}

func main() {

	port := "443"
	args := os.Args
	// fmt.Println(len(args))
	if !fileutil.HasStdin() && len(args) != 2 {
		fmt.Println("Please provide one file with list of IPs")
	} else {
		ips := getips(args)
		for _, ip := range ips {

			if strings.Count(ip, ":") == 0 {
				var ip_port string = ip + ":" + port
				hostname := gethostname(ip_port)
				fmt.Println(ip, strings.ToLower(hostname))
			} else if strings.Count(ip, ":") == 1 && !strings.Contains(ip, "https") {
				string_split := strings.Split(ip, ":")
				var ip_port string = string_split[0] + ":" + string_split[1]
				hostname := gethostname(ip_port)
				fmt.Println(ip, strings.ToLower(hostname))
			} else if strings.Count(ip, ":") == 1 && strings.Contains(ip, "https") {
				string_split := strings.Split(ip, "//")
				var ip_port string = string_split[1] + ":" + port
				hostname := gethostname(ip_port)
				fmt.Println(ip, strings.ToLower(hostname))
			} else if strings.Count(ip, ":") == 2 && strings.Contains(ip, "https") {
				string_split := strings.Split(ip, ":")
				var ip_port string = strings.Replace(string_split[1], "//", "", -1) + ":" + string_split[2]
				hostname := gethostname(ip_port)
				fmt.Println(ip, strings.ToLower(hostname))
			}
			// } else if strings.Contains(ip, "http://") {
			// 	fmt.Println("HTTP url provided. HTTPS needed")
			// } else {
			// 	fmt.Println("IP address format is wrong")
			// }

		}
	}

}