# Fly2User

Fly2User is a self-contained user manager designed for managing users on V2Ray servers. It provides APIs for adding, removing, purging, and counting users while supporting multiple upstream servers. It is optimized to detect changes such as upstream server upscale, downscale, and restarts automatically.

## Features

- **User Management**: Add, remove, purge, and count users on V2Ray servers.
- **Multiple Upstream Servers**: Supports managing multiple upstream V2Ray servers, even those with the same hostname.
- **Auto Detection**: Automatically detects upstream server upscale, downscale, and restarts.

## Configuration

Fly2User is configured using environment variables. Below are the configuration keys:

| Config Key              | Description                                                                                | Default       | Mandatory |
|-------------------------|--------------------------------------------------------------------------------------------|---------------|-----------|
| `SERVER__AUTH`           | Authentication for the Fly2User server in the format `username:password`. Not mandatory.   | None          | No        |
| `SERVER__LISTEN`         | The address Fly2User will listen on (e.g., `127.0.0.1:3000`).                              | `:3000`       | No        |
| `SUPERVISOR__USER_DB`    | Path to store the user database (e.g., `path/user.sqlite`).                                | None          | Yes       |
| `SUPERVISOR__INTERVAL`   | Interval in seconds to check for upstream server restarts.                                 | `60`          | No        |
| `UPSTREAM__ADDRESS`      | Comma-separated list of upstream V2Ray API listen addresses in the format: `grpc://SERVERTYPE@ADDRESS:PORT`, where `SERVERTYPE` is one of `v2ray` or `v2fly`. | None          | Yes       |
| `UPSTREAM__INBOUNDS`     | Comma-separated list of inbound configurations in the format `TAG:TYPE`, where `TYPE` is one of `VMESS`, `VLESS`, or `TROJAN`. | None          | Yes       |
| `LOG_LEVEL`              | Log level for the Fly2User server.                                                         | `WARNING`     | No        |

### Example Configuration

```bash
export SERVER__AUTH=username:password
export SERVER__LISTEN=127.0.0.1:3000
export SUPERVISOR__USER_DB=/path/to/user.sqlite
export SUPERVISOR__INTERVAL=30
export UPSTREAM__ADDRESS=grpc://v2ray@127.0.0.1:10085,grpc://v2fly@127.0.0.1:10086
export UPSTREAM__INBOUNDS=vmess_inbound:VMESS,vless_inbound:VLESS
export LOG_LEVEL=DEBUG
```

## API Endpoints

Fly2User exposes a simple REST API for managing users:

### Add User

- **Endpoint**: `POST /user/`
- **Description**: Add a new user to the system.
- **Input**:
  ```json
  {
    "uuid": "user-uuid",
    "email": "user@example.com",
    "level": 1
  }
  ```

### Remove User

- **Endpoint**: `DELETE /user/`
- **Description**: Remove an existing user by email.
- **Input**:
  ```json
  {
    "email": "user@example.com"
  }
  ```

### Purge All Users

- **Endpoint**: `GET /user/flush`
- **Description**: Purge all users from the system.

### Count Users

- **Endpoint**: `GET /user/count`
- **Description**: Get the count of all users in the system.


## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.