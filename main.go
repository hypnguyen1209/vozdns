package main

import (
	"flag"
	"fmt"
)

type ClientConfig struct {
	PrivateKey string `json:"privatekey"`
	PublicKey  string `json:"publickey"`
	Domain     string `json:"domain"`
	ProxySSL   bool   `json:"proxy_ssl"`
}

type ServerConfig struct {
	PrivateKey string `json:"privatekey"`
	PublicKey  string `json:"publickey"`
	Listen     string `json:"listen"`
	AuthEmail  string `json:"auth_email"`
	AuthKey    string `json:"auth_key"`
	ZoneID     string `json:"zone_id"`
}

type VerifyRequest struct {
	Domain   string `json:"domain"`
	ProxySSL bool   `json:"proxy_ssl"`
	IP       string `json:"ip"`
}

type VerifyResponse struct {
	Domain    string `json:"domain"`
	ProxySSL  bool   `json:"proxy_ssl"`
	IP        string `json:"ip"`
	PublicKey string `json:"publickey"`
}

type RegisterRequest struct {
	EncryptedData string `json:"encrypted_data"`
}

type ServerInfo struct {
	Server string `json:"server"`
}

type AuthorizedDomain struct {
	Domain    string `json:"domain"`
	PublicKey string `json:"publickey"`
}

func main() {
	var (
		generate       = flag.Bool("generate", false, "Generate client config")
		generateServer = flag.Bool("generate-server", false, "Generate server config")
		start          = flag.Bool("start", false, "Start client")
		server         = flag.Bool("server", false, "Start server")
		domain         = flag.String("domain", "example.vozdns.vn", "Domain for client config")
	)

	flag.Parse()

	switch {
	case *generate:
		generateClientConfig(*domain)
	case *generateServer:
		generateServerConfig()
	case *start:
		startClient()
	case *server:
		startServer()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ./vozdns -generate [-domain <domain>]  # Generate client config")
	fmt.Println("  ./vozdns -generate-server              # Generate server config")
	fmt.Println("  ./vozdns -start                        # Start client")
	fmt.Println("  ./vozdns -server                       # Start server")
	fmt.Println("")
	flag.PrintDefaults()
}
