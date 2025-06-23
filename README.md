# Go Fiber Template

This project provides a basic setup for a Go Fiber REST API.

## Chạy kiểm thử

Sử dụng lệnh sau để chạy toàn bộ test:

```bash
go test ./...
```

Kết quả sẽ hiển thị các test thành công hoặc thất bại trên terminal.

## Khởi chạy server

Tạo file `.env` từ mẫu `env` và chỉnh sửa thông tin kết nối nếu cần. Sau đó có thể chạy server bằng lệnh:

```bash
go run main.go
```

Hoặc sử dụng Docker Compose:

```bash
docker-compose up
```

Mặc định ứng dụng lắng nghe tại `http://localhost:4000`.

## Danh sách API chính

| Phương thức | Đường dẫn              | Ghi chú                           |
|-------------|-----------------------|----------------------------------|
| POST        | `/login`              | Đăng nhập nhận JWT               |
| GET         | `/test`               | Kiểm tra server                  |
| GET         | `/api/test2`          | Cần JWT                          |
| GET         | `/api/me`             | Thông tin người dùng hiện tại    |
| PUT         | `/api/users/password` | Đổi mật khẩu người dùng          |
| PUT         | `/api/presigned_url`  | Tạo URL upload tạm thời          |
| POST        | `/api/users`          | Chỉ admin, tạo người dùng mới    |
| GET         | `/api/users`          | Chỉ admin, danh sách người dùng  |

Sau khi đăng nhập thành công, thêm header `Authorization: Bearer <token>` khi gọi các đường dẫn `/api/*`.

Ví dụ đăng nhập với tài khoản mặc định:

```bash
curl -X POST http://localhost:4000/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```
