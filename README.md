# VozDNS - Dynamic DNS Client

VozDNS is a secure dynamic DNS client that automatically updates your subdomain's DNS record when your IP address changes. Perfect for home servers, development environments, or any service that needs a stable domain name with a dynamic IP.

🇻🇳 Vietnamese version: [README_VI.md](README_VI.md)

## 📋 Prerequisites

Before using VozDNS, you need:
1. A subdomain registered in the system (see [Getting a Subdomain](#getting-a-subdomain))
2. The VozDNS client binary for your platform

## 🔗 Getting a Subdomain

To get a subdomain (e.g., `yourname.vozdns.vn`), you need to submit a pull request:

1. **Fork this repository** on GitHub
2. **Edit the `subdomain.json` file** and add your entry:
   ```json
   {
     "domain": "yourname.vozdns.vn",
     "publickey": "your-public-key-will-be-generated"
   }
   ```
3. **Create a pull request** with your subdomain request
4. **Wait for approval** - once merged, your subdomain will be active

> **Note**: You'll generate your public key in the next step, then update your pull request with the actual key.

## 📥 Installation

### Option 1: Download Pre-built Binary
Download the latest binary for your platform from the [Releases](https://github.com/hypnguyen1209/vozdns/releases) page.

Available platforms:
- Linux (amd64, arm64, arm, 386)
- Windows (amd64, 386)
- macOS (amd64, arm64)
- FreeBSD (amd64)

### Option 2: Docker Container
```bash
# Pull the latest image
docker pull ghcr.io/hypnguyen1209/vozdns:latest

# Run as client (with config volume)
docker run -d --name vozdns-client \
  -v /path/to/config:/home/appuser/.vozdns:ro \
  ghcr.io/hypnguyen1209/vozdns:latest -start

# Or use docker-compose
docker-compose up -d vozdns-client
```

### Option 3: Build from Source
```bash
# Clone the repository
git clone https://github.com/hypnguyen1209/vozdns.git
cd vozdns

# Build the binary
go build -o vozdns

# Make it executable (Linux/macOS)
chmod +x vozdns
```

## ⚙️ Setup

### Step 1: Generate Client Configuration

```bash
# Generate configuration for your subdomain
./vozdns -generate -domain yourname.vozdns.vn
```

This creates a configuration file at:
- **Linux/macOS**: `$HOME/.vozdns/config.json`
- **Windows**: `%USERPROFILE%\.vozdns\config.json`

The generated config contains:
```json
{
  "privatekey": "<your-private-key>",
  "publickey": "<your-public-key>",
  "domain": "yourname.vozdns.vn",
  "proxy_ssl": false
}
```

### Step 2: Submit Your Public Key

1. **Copy your public key** from the generated config file
2. **Update your pull request** (from the subdomain registration step) with your actual public key
3. **Wait for the pull request to be merged**

### Step 3: Start the Client

Once your subdomain is approved and merged:

```bash
./vozdns -start
```

## 🔄 How It Works

1. **IP Detection**: Client detects your current public IP address
2. **Server Discovery**: Fetches server information from `https://vozdns.vn/server.json`
3. **Authorization**: Server verifies your domain against `https://vozdns.vn/subdomain.json`
4. **Secure Communication**: All data is encrypted using your ECC key pair
5. **DNS Update**: If your IP changed, updates the DNS record via Cloudflare
6. **Repeat**: Process repeats every 10 minutes

## 🔧 Configuration Options

### Client Configuration File

Located at `$HOME/.vozdns/config.json`:

| Field | Description | Default |
|-------|-------------|---------|
| `privatekey` | Your private key (keep secret!) | Generated |
| `publickey` | Your public key (shared with server) | Generated |
| `domain` | Your subdomain | Required |
| `proxy_ssl` | Enable Cloudflare proxy | `false` |

### Command Line Options

```bash
./vozdns -help
```

Available flags:
- `-generate`: Generate client configuration
- `-domain string`: Specify domain for config generation
- `-start`: Start the client
- `-server`: Start server (admin only)
- `-generate-server`: Generate server config (admin only)

## 📊 Monitoring and Logs

The client outputs detailed logs showing:
- Current IP address detection
- Server communication status
- DNS record updates
- Error messages and troubleshooting info

Example output:
```
Starting VozDNS client...
Loaded config for domain: yourname.vozdns.vn
[2025-07-21 10:22:19] Starting client cycle...
Public IP: 203.0.113.42
Server: server.vozdns.vn:9000
Verification successful, server public key received
Registration successful
VozDNS client started. Press Ctrl+C to stop.
```

## 🚨 Troubleshooting

### Common Issues

**"Domain not authorized"**
- Your subdomain hasn't been approved yet
- Check if your pull request was merged
- Verify your domain name matches exactly

**"Config file not found"**
- Run `./vozdns -generate -domain yourname.vozdns.vn` first

**"Connection failed"**
- Check your internet connection
- Server might be temporarily unavailable

**"Failed to get public IP"**
- Network connectivity issues
- Firewall blocking outbound connections

### Getting Help

1. Check the logs for detailed error messages
2. Verify your configuration file is correct
3. Ensure your subdomain was approved and merged
4. Create an issue on GitHub with logs and configuration details

## 🔐 Security Notes

- **Keep your private key secure** - never share it
- Only your public key is stored in the public subdomain registry
- All communication with the server is encrypted
- DNS updates require valid domain authorization

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes

## 📞 Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/hypnguyen1209/vozdns/issues)
- **Documentation**: Check this README for detailed instructions


### README create by ChatGPT ♥️