package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)


func getConfigDir() (string, error) {
	var homeDir string
	var err error

	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}

	if homeDir == "" {
		return "", fmt.Errorf("unable to determine home directory")
	}

	configDir := filepath.Join(homeDir, ".vozdns")
	
	
	if err = os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	return configDir, nil
}


func generateClientConfig(domain string) {
	fmt.Printf("Generating client config for domain: %s\n", domain)

	
	privateKey, publicKey, err := generateECCKeyPair()
	if err != nil {
		fmt.Printf("Error generating key pair: %v\n", err)
		return
	}

	
	privateKeyStr, err := encodePrivateKey(privateKey)
	if err != nil {
		fmt.Printf("Error encoding private key: %v\n", err)
		return
	}

	publicKeyStr, err := encodePublicKey(publicKey)
	if err != nil {
		fmt.Printf("Error encoding public key: %v\n", err)
		return
	}

	
	config := ClientConfig{
		PrivateKey: privateKeyStr,
		PublicKey:  publicKeyStr,
		Domain:     domain,
		ProxySSL:   false,
	}

	
	configDir, err := getConfigDir()
	if err != nil {
		fmt.Printf("Error getting config directory: %v\n", err)
		return
	}

	configPath := filepath.Join(configDir, "config.json")
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	err = os.WriteFile(configPath, configData, 0600)
	if err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

	fmt.Printf("Client config generated successfully at: %s\n", configPath)
	fmt.Printf("Public key: %s\n", publicKeyStr)
}


func generateServerConfig() {
	fmt.Println("Generating server config...")

	
	privateKey, publicKey, err := generateECCKeyPair()
	if err != nil {
		fmt.Printf("Error generating key pair: %v\n", err)
		return
	}

	
	privateKeyStr, err := encodePrivateKey(privateKey)
	if err != nil {
		fmt.Printf("Error encoding private key: %v\n", err)
		return
	}

	publicKeyStr, err := encodePublicKey(publicKey)
	if err != nil {
		fmt.Printf("Error encoding public key: %v\n", err)
		return
	}

	
	config := ServerConfig{
		PrivateKey: privateKeyStr,
		PublicKey:  publicKeyStr,
		Listen:     ":8080",
		AuthEmail:  "(The email used to login 'https://dash.cloudflare.com')",
		AuthKey:    "(Your API Token)",
		ZoneID:     "(Can be found in the \"Overview\" tab of your domain)",
	}

	
	configPath := "./config.json"
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	err = os.WriteFile(configPath, configData, 0600)
	if err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

	fmt.Printf("Server config generated successfully at: %s\n", configPath)
	fmt.Println("Please edit the config file and update the Cloudflare credentials.")
}


func loadClientConfig() (*ClientConfig, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found. Run './vozdns -generate' first")
	}

	var config ClientConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid config file: %v", err)
	}

	return &config, nil
}


func loadServerConfig() (*ServerConfig, error) {
	configPath := "./config.json"
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found. Run './vozdns -generate-server' first")
	}

	var config ServerConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid config file: %v", err)
	}

	return &config, nil
}
