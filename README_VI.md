# VozDNS - Ứng dụng DNS Động

VozDNS là một ứng dụng khách DNS động bảo mật, tự động cập nhật bản ghi DNS của subdomain khi địa chỉ IP thay đổi. Hoàn hảo cho máy chủ gia đình, môi trường phát triển, hoặc bất kỳ dịch vụ nào cần tên miền ổn định với IP động.

🇺🇸 Phiên bản tiếng Anh: [README.md](README.md)

## 📋 Yêu cầu

Trước khi sử dụng VozDNS, bạn cần có:
1. Một subdomain được đăng ký trong hệ thống (xem phần [Đăng ký Subdomain](#đăng-ký-subdomain))
2. File thực thi VozDNS client cho hệ điều hành của bạn

## 🔗 Đăng ký Subdomain

Để có được một subdomain (ví dụ: `yourname.vozdns.vn`), bạn cần gửi pull request:

1. **Fork repository này** trên GitHub
2. **Chỉnh sửa file `subdomain.json`** và thêm thông tin của bạn:
   ```json
   {
     "domain": "yourname.vozdns.vn",
     "publickey": "your-public-key-will-be-generated"
   }
   ```
3. **Tạo pull request** với yêu cầu đăng ký subdomain
4. **Chờ phê duyệt** - sau khi được merge, subdomain của bạn sẽ hoạt động

> **Lưu ý**: Bạn sẽ tạo public key ở bước tiếp theo, sau đó cập nhật pull request với key thực tế.

## 📥 Cài đặt

### Cách 1: Tải file thực thi có sẵn
Tải binary mới nhất [Releases](https://github.com/hypnguyen1209/vozdns/releases).

### Cách 2: Biên dịch từ mã nguồn
```bash
# Clone repository
git clone https://github.com/hypnguyen1209/vozdns.git
cd vozdns

# Biên dịch binary
go build -o vozdns

# Cấp quyền thực thi (Linux/macOS)
chmod +x vozdns
```

## ⚙️ Thiết lập

### Bước 1: Tạo cấu hình Client

```bash
# Tạo cấu hình cho subdomain của bạn
./vozdns -generate -domain yourname.vozdns.vn
```

Lệnh này sẽ tạo file cấu hình tại:
- **Linux/macOS**: `$HOME/.vozdns/config.json`
- **Windows**: `%USERPROFILE%\.vozdns\config.json`

Nội dung file cấu hình được tạo:
```json
{
  "privatekey": "<private-key-của-bạn>",
  "publickey": "<public-key-của-bạn>",
  "domain": "yourname.vozdns.vn",
  "proxy_ssl": false
}
```

### Bước 2: Gửi Public Key

1. **Sao chép public key** từ file cấu hình vừa tạo
2. **Cập nhật pull request** (từ bước đăng ký subdomain) với public key thực tế
3. **Chờ pull request được merge**

### Bước 3: Khởi động Client

Khi subdomain đã được phê duyệt và merge:

```bash
./vozdns -start
```

## 🔄 Cách thức hoạt động

1. **Phát hiện IP**: Client tự động phát hiện địa chỉ IP công khai hiện tại
2. **Kết nối Server**: Lấy thông tin server từ `https://vozdns.vn/server.json`
3. **Xác thực**: Server xác minh quyền sở hữu domain qua `https://vozdns.vn/subdomain.json`
4. **Mã hóa giao tiếp**: Tất cả dữ liệu được mã hóa bằng cặp khóa ECC
5. **Cập nhật DNS**: Nếu IP thay đổi, hệ thống cập nhật bản ghi DNS qua Cloudflare
6. **Lặp lại**: Quy trình được lặp lại mỗi 10 phút

## 🔧 Tùy chọn cấu hình

### File cấu hình Client

Tệp cấu hình đặt tại `$HOME/.vozdns/config.json`:

| Trường | Mô tả | Giá trị mặc định |
|--------|-------|------------------|
| `privatekey` | Khóa riêng tư (bảo mật tuyệt đối!) | Được tạo tự động |
| `publickey` | Khóa công khai (chia sẻ với server) | Được tạo tự động |
| `domain` | Subdomain của bạn | Bắt buộc |
| `proxy_ssl` | Bật Cloudflare proxy | `false` |

### Tùy chọn dòng lệnh

```bash
./vozdns -help
```

Arguments:
- `-generate`: Tạo cấu hình client
- `-domain string`: Chỉ định domain cho việc tạo cấu hình
- `-start`: Khởi động client
- `-server`: Khởi động server (chỉ dành cho quản trị viên)
- `-generate-server`: Tạo cấu hình server (chỉ dành cho quản trị viên)

## 📊 Giám sát và Nhật ký

Client xuất ra nhật ký chi tiết hiển thị:
- Phát hiện địa chỉ IP hiện tại
- Trạng thái giao tiếp với server
- Cập nhật bản ghi DNS
- Thông báo lỗi và thông tin khắc phục

Ví dụ kết quả đầu ra:
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

## 🚨 Khắc phục sự cố

### Các vấn đề thường gặp

**"Domain not authorized" (Domain không được phép)**
- Subdomain của bạn chưa được phê duyệt
- Kiểm tra xem pull request đã được merge chưa
- Xác minh tên domain khớp chính xác

**"Config file not found" (Không tìm thấy file cấu hình)**
- Chạy lệnh `./vozdns -generate -domain yourname.vozdns.vn` trước

**"Connection failed" (Kết nối thất bại)**
- Kiểm tra kết nối internet
- Server có thể tạm thời không khả dụng

**"Failed to get public IP" (Không lấy được IP công khai)**
- Có vấn đề với kết nối mạng
- Firewall có thể đang chặn kết nối ra ngoài

### Nhận trợ giúp

1. Kiểm tra nhật ký để biết thông báo lỗi chi tiết
2. Xác minh file cấu hình đúng định dạng
3. Đảm bảo subdomain đã được phê duyệt và merge
4. Tạo issue trên GitHub kèm nhật ký và chi tiết cấu hình

## 🔐 Lưu ý bảo mật

- **Bảo vệ khóa riêng tư** - không bao giờ chia sẻ với ai
- Chỉ có public key được lưu trữ trong registry subdomain công khai
- Tất cả giao tiếp với server đều được mã hóa
- Việc cập nhật DNS yêu cầu xác thực domain hợp lệ

## 📄 Giấy phép

Giấy phép MIT - xem file [LICENSE](LICENSE) để biết chi tiết.

## 🤝 Đóng góp

Chúng tôi hoan nghênh các đóng góp! Vui lòng:
1. Fork repository
2. Tạo feature branch
3. Thực hiện các thay đổi
4. Gửi pull request

## 📞 Hỗ trợ

- **GitHub Issues**: [Báo cáo lỗi hoặc yêu cầu tính năng mới](https://github.com/hypnguyen1209/vozdns/issues)
- **Tài liệu**: Tham khảo README này để biết hướng dẫn chi tiết


### README create by ChatGPT ♥️