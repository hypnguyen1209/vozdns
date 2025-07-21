package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getPublicIP() (string, error) {
	resp, err := http.Get("https://icanhazip.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := string(bytes.TrimSpace(body))
	return ip, nil
}

func getServerInfo() (*ServerInfo, error) {
	resp, err := http.Get("https://vozdns.vn/server.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serverInfo ServerInfo
	err = json.NewDecoder(resp.Body).Decode(&serverInfo)
	if err != nil {
		return nil, err
	}

	return &serverInfo, nil
}

func verifyWithServer(serverURL string, config *ClientConfig, ip string) (*VerifyResponse, error) {
	verifyReq := VerifyRequest{
		Domain:   config.Domain,
		ProxySSL: config.ProxySSL,
		IP:       ip,
	}

	reqData, err := json.Marshal(verifyReq)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/verify", serverURL), "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	encryptedData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	clientPrivateKey, err := decodePrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding private key: %v", err)
	}

	decryptedData, err := decryptWithPrivateKey(string(encryptedData), clientPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting response: %v", err)
	}

	var verifyResp VerifyResponse
	err = json.Unmarshal(decryptedData, &verifyResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &verifyResp, nil
}

func registerWithServer(serverURL string, config *ClientConfig, ip string, serverPublicKey string) error {

	registerData := map[string]interface{}{
		"domain":    config.Domain,
		"proxy_ssl": config.ProxySSL,
		"ip":        ip,
	}

	registerJSON, err := json.Marshal(registerData)
	if err != nil {
		return err
	}

	serverPubKey, err := decodePublicKey(serverPublicKey)
	if err != nil {
		return fmt.Errorf("error decoding server public key: %v", err)
	}

	encryptedData, err := encryptWithPublicKey(registerJSON, serverPubKey)
	if err != nil {
		return fmt.Errorf("error encrypting data: %v", err)
	}

	reqData, err := json.Marshal(map[string]string{"encrypted_data": encryptedData})
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/register", serverURL), "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func startClient() {
	fmt.Println("Starting VozDNS client...")

	config, err := loadClientConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Loaded config for domain: %s\n", config.Domain)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	runClientCycle(config)

	fmt.Println("VozDNS client started. Press Ctrl+C to stop.")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client context cancelled, shutting down...")
			return
		case <-sigChan:
			fmt.Println("\nReceived shutdown signal, stopping VozDNS client...")
			cancel()
			return
		case <-ticker.C:
			runClientCycle(config)
		}
	}
}

func runClientCycle(config *ClientConfig) {
	fmt.Printf("[%s] Starting client cycle...\n", time.Now().Format("2006-01-02 15:04:05"))

	ip, err := getPublicIP()
	if err != nil {
		fmt.Printf("Error getting public IP: %v\n", err)
		return
	}
	fmt.Printf("Public IP: %s\n", ip)

	serverInfo, err := getServerInfo()
	if err != nil {
		fmt.Printf("Error getting server info: %v\n", err)
		return
	}
	fmt.Printf("Server: %s\n", serverInfo.Server)

	verifyResp, err := verifyWithServer(serverInfo.Server, config, ip)
	if err != nil {
		fmt.Printf("Error verifying with server: %v\n", err)
		return
	}
	fmt.Printf("Verification successful, server public key received\n")

	err = registerWithServer(serverInfo.Server, config, ip, verifyResp.PublicKey)
	if err != nil {
		fmt.Printf("Error registering with server: %v\n", err)
		return
	}
	fmt.Printf("Registration successful\n")
}
