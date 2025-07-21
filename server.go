package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hypnguyen1209/ming/v2"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)


type CloudflareRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}


func isAuthorizedDomain(domain string) (bool, string, error) {
	resp, err := http.Get("https://vozdns.vn/subdomain.json")
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var authorizedDomains []AuthorizedDomain
	err = json.NewDecoder(resp.Body).Decode(&authorizedDomains)
	if err != nil {
		return false, "", err
	}

	for _, authDomain := range authorizedDomains {
		if authDomain.Domain == domain {
			return true, authDomain.PublicKey, nil
		}
	}

	return false, "", nil
}


func getCloudflareRecord(config *ServerConfig, domain string) (*CloudflareRecord, error) {
	
	if config.AuthKey == "(Your API Token)" || config.ZoneID == "(Can be found in the \"Overview\" tab of your domain)" {
		fmt.Printf("Using test credentials - simulating DNS record check for %s\n", domain)
		
		return &CloudflareRecord{
			ID:      "test-record-id",
			Type:    "A",
			Name:    domain,
			Content: "0.0.0.0", 
			Proxied: false,
		}, nil
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s&type=A", config.ZoneID, domain)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudflare API returned status %d", resp.StatusCode)
	}

	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	
	if !gjson.GetBytes(body, "success").Bool() {
		errors := gjson.GetBytes(body, "errors").Array()
		return nil, fmt.Errorf("cloudflare API error: %v", errors)
	}

	
	records := gjson.GetBytes(body, "result").Array()
	if len(records) == 0 {
		return nil, nil 
	}

	
	firstRecord := records[0]
	record := &CloudflareRecord{
		ID:      firstRecord.Get("id").String(),
		Type:    firstRecord.Get("type").String(),
		Name:    firstRecord.Get("name").String(),
		Content: firstRecord.Get("content").String(),
		Proxied: firstRecord.Get("proxied").Bool(),
	}

	return record, nil
}


func updateCloudflareRecord(config *ServerConfig, domain, ip string, proxied bool) error {
	
	if config.AuthKey == "(Your API Token)" || config.ZoneID == "(Can be found in the \"Overview\" tab of your domain)" {
		fmt.Printf("Using test credentials - simulating DNS update for %s -> %s (proxied: %v)\n", domain, ip, proxied)
		return nil 
	}

	
	existingRecord, err := getCloudflareRecord(config, domain)
	if err != nil {
		return err
	}

	var url string
	var method string

	recordData := map[string]interface{}{
		"type":    "A",
		"name":    domain,
		"content": ip,
		"proxied": proxied,
	}

	if existingRecord != nil {
		
		url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", config.ZoneID, existingRecord.ID)
		method = "PUT"
	} else {
		
		url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", config.ZoneID)
		method = "POST"
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+config.AuthKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	
	if !gjson.GetBytes(body, "success").Bool() {
		errors := gjson.GetBytes(body, "errors").Array()
		return fmt.Errorf("cloudflare API error: %v", errors)
	}

	return nil
}


func startServer() {
	fmt.Println("Starting VozDNS server...")

	
	config, err := loadServerConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Server starting on %s\n", config.Listen)

	
	router := ming.New()

	
	router.Post("/verify", func(ctx *fasthttp.RequestCtx) {
		var verifyReq VerifyRequest
		if err := json.Unmarshal(ctx.PostBody(), &verifyReq); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Invalid request"}`)
			return
		}

		
		verifyResp := VerifyResponse{
			Domain:    verifyReq.Domain,
			ProxySSL:  verifyReq.ProxySSL,
			IP:        verifyReq.IP,
			PublicKey: config.PublicKey,
		}

		
		authorized, clientPublicKey, err := isAuthorizedDomain(verifyReq.Domain)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Error checking authorization"}`)
			return
		}
		
		if !authorized {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Domain not authorized"}`)
			return
		}

		
		respData, err := json.Marshal(verifyResp)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Failed to marshal response"}`)
			return
		}

		clientPubKey, err := decodePublicKey(clientPublicKey)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Invalid client public key"}`)
			return
		}

		encryptedResp, err := encryptWithPublicKey(respData, clientPubKey)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Failed to encrypt response"}`)
			return
		}

		ctx.SetContentType("text/plain")
		ctx.WriteString(encryptedResp)
	})

	
	router.Post("/register", func(ctx *fasthttp.RequestCtx) {
		var registerReq RegisterRequest
		if err := json.Unmarshal(ctx.PostBody(), &registerReq); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Invalid request"}`)
			return
		}

		
		serverPrivateKey, err := decodePrivateKey(config.PrivateKey)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Server key error"}`)
			return
		}

		decryptedData, err := decryptWithPrivateKey(registerReq.EncryptedData, serverPrivateKey)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Failed to decrypt request"}`)
			return
		}

		var registerData VerifyRequest
		err = json.Unmarshal(decryptedData, &registerData)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Invalid decrypted data"}`)
			return
		}

		
		authorized, _, err := isAuthorizedDomain(registerData.Domain)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Error checking authorization"}`)
			return
		}
		if !authorized {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error": "Domain not authorized"}`)
			return
		}

		
		currentRecord, err := getCloudflareRecord(config, registerData.Domain)
		if err != nil {
			fmt.Printf("Error checking DNS record for %s: %v\n", registerData.Domain, err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(fmt.Sprintf(`{"error": "Failed to check DNS record: %v"}`, err))
			return
		}

		
		if currentRecord == nil || currentRecord.Content != registerData.IP {
			err = updateCloudflareRecord(config, registerData.Domain, registerData.IP, registerData.ProxySSL)
			if err != nil {
				fmt.Printf("Error updating DNS record for %s -> %s: %v\n", registerData.Domain, registerData.IP, err)
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetContentType("application/json")
				ctx.WriteString(fmt.Sprintf(`{"error": "Failed to update DNS record: %v"}`, err))
				return
			}
			fmt.Printf("Updated DNS record: %s -> %s\n", registerData.Domain, registerData.IP)
		} else {
			fmt.Printf("DNS record already up to date: %s -> %s\n", registerData.Domain, registerData.IP)
		}

		ctx.SetContentType("application/json")
		ctx.WriteString(`{"status": "success"}`)
	})

	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	
	go func() {
		fmt.Printf("Server listening on %s\n", config.Listen)
		router.Run(config.Listen)
	}()

	
	<-quit
	fmt.Println("\nShutting down server...")
	fmt.Println("Server stopped gracefully")
}
