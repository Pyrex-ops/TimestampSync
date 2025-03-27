# MPV Timestamp Sync Plugin

This project provides seamless media timestamp synchronization for **MPV** using a **self-hosted Golang web server**.

It consists of:
- A **Lua script** for MPV that uploads and retrieves playback timestamps from the server.
- A **Golang web server** that stores timestamps in an SQLite database.
- **Automatic garbage collection** to remove old timestamp entries.
- **Docker support** for easy deployment.

With this setup, you can **resume playback across multiple MPV instances** without manual tracking.

---

## Features

✅ **Automatic timestamp saving** – Saves progress when MPV exits.  
✅ **Automatic timestamp restoration** – Resumes from the last saved position.  
✅ **Self-hosted** – No third-party services required.  
✅ **SQLite storage** – Lightweight and efficient.  
✅ **Basic authentication** – Authenticated API access.  
✅ **Garbage collection** – Cleans up old timestamps periodically.  
✅ **Docker support** – Easy deployment with a lightweight container.  

---

## ⚠️ Security Warning

> [!WARNING]
> **The web server must be deployed behind an HTTPS reverse proxy.**  
> Currently, **HTTP Basic Authentication** is used, which is **not secure over plain HTTP**. Use **Nginx, Caddy, or Traefik** as a reverse proxy with **HTTPS enabled** to protect credentials.

---

## Installation

### 1. Install MPV, Lua and curl

Ensure MPV is installed with Lua scripting support.
You need to have curl installed as well.

### 2. Install the Lua Script

1. Copy `upload_time.lua` to your MPV scripts directory:
   - **Linux/macOS**: `~/.config/mpv/scripts/`
   - **Windows**: `%APPDATA%\mpv\scripts\`
     - On Windows it may be necessary to indicate the ".exe" extension (curl.exe) 

### 3. Set Up the Golang Server

#### Option A: Run natively

1. Install Go if you haven’t already.
2. Clone the repository and navigate into it:
   ```sh
   git clone https://github.com/Pyrex-ops/TimestampSync.git
   cd TimestampSync
   ```  
3. Build and run the server:
   ```sh
   go build -o timestamp-server  
   ./timestamp-server  
   ```  

#### Option B: Run with Docker

1. Ensure **Docker** is installed.
2. Clone the repository:
   ```sh
   git clone https://github.com/Pyrex-ops/TimestampSync.git
   cd TimestampSync
   ```  
3. Build and run the Docker container:
   ```sh
   docker build -t mpv-timestamp-server .  
   docker run -d -e TZ="Europe/Rome" -p 8080:8080 --name mpv-server mpv-timestamp-server  
   ```  

---

## Configuration

Modify the Lua script (`upload_time.lua`) to match your server’s address and credentials:

```lua
local api_url = "http://localhost:8080/timestamps"
local username = "admin"
local password = "change_password"
```

If your server runs on a different machine, update it accordingly (e.g., `"http://192.168.1.100:8080"`).

---

## Usage

1. **Start the server** before using MPV.
2. **Open a media file** in MPV – it will check for a saved timestamp and resume from there.
3. **Exit MPV** – the current timestamp is uploaded to the server.
4. The next time you play the same file, it will resume automatically.

---

## API Endpoints

The Golang server provides a simple API with **basic authentication**.

### **Get all timestamps**
```http
GET /timestamps  
Authorization: Basic <base64-encoded-credentials>
Content-Type: application/json  
```
**Response:**
```json
[
  {
    "ID": -8465237232429249018,
    "Name": "file1.aac",
    "Seconds": 1206,
    "Timestamp": "2025-01-03T12:45:33+08:00"
  },
  {
    "ID": -7044684209813081264,
    "Name": "file2.mp3",
    "Seconds": 135,
    "Timestamp": "2014-07-21T21:01:48+08:00"
  }
]
```

### **Get timestamp by media name**
```http
GET /timestamps/{name}  
Authorization: Basic <base64-encoded-credentials>
Content-Type: application/json  
```
**Response:**
```json
{"ID":7278257633394833063,"Name":"file3.mp3","Seconds":100,"Timestamp":"2011-08-13T16:09:50+08:00"}
```

### **Save (or update) a timestamp**
```http
POST /timestamps  
Authorization: Basic <base64-encoded-credentials>  
Content-Type: application/json  

{
  "name": "movie.mp4",
  "seconds": 123
}
```
**Response:**
```json
{ "message": "Timestamp saved successfully" }
```

### **Delete a timestamp by ID**
```http
DELETE /timestamps/{id}  
Authorization: Basic <base64-encoded-credentials>
```
**Response:**
```json
{ "message": "Timestamp deleted successfully" }
```

---

## Automatic Garbage Collection

To **prevent database bloat**, a scheduled **garbage collector** runs **every two weeks at 4:00 AM** to remove old timestamp entries.

- Uses the [gocron](https://github.com/go-co-op/gocron) library.
- Cleans up expired or unnecessary entries automatically.
- Logs the next scheduled cleanup time.

If you need to adjust the cleanup schedule, set the `GC_SCHEDULE` variable in the `.env` file or `export` it before starting the server.  
Example "At 14:15 on day-of-month 1":
```shell
export GC_SCHEDULE="15 14 1 * *"
```

---

## Running with Docker

A **Dockerfile** is included for easy deployment.

### **Build and Run the Container**

```sh
docker build -t mpv-timestamp-server .  
docker run -d -e TZ="Europe/Rome" -p 8080:8080 --name mpv-server mpv-timestamp-server  
```

### **Dockerfile Overview**

The project uses a **multi-stage build**:

#### **Build Stage** (Golang)
- Uses the latest Go image.
- Downloads dependencies and compiles the binary with optimizations (`-w -s -trimpath`).

#### **Runtime Stage** (Debian Slim)
- Uses a minimal Debian image for a small container size.
- Copies the compiled binary from the build stage.
- Runs the server on container startup.

---

## Notes

- Timestamps are **matched by filename** – if the filename changes, the timestamp won’t be found.
- The API requires **basic authentication** – update credentials by setting env variables `BASIC_AUTH_USER` and `BASIC_AUTH_PASS`.
- Ensure the server is **always running** when using MPV for proper timestamp syncing.
- **Reverse Proxy Required:** Since HTTP Basic Authentication is used, deploy the web server **behind an HTTPS reverse proxy**.

---

## Future Enhancements

🔹 **Support for multiple users** with individual timestamps  
🔹 **More advanced filename matching** and metadata storage  
🔹 **VLC** client

---

## License

This project is licensed under the **GPLv3 Licence**.

---

## Contributing

Pull requests and improvements are welcome! Feel free to submit issues or feature requests.

---

Enjoy seamless playback synchronization across devices with **MPV Timestamp Sync**! 🎥