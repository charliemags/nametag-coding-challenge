# Self-Updating Go Application

### This repository contains a proof-of-concept self-updating Go program that works (in principle) on Windows, macOS, and Linux.
**A minimal server is provided, which:**

- Hosts metadata describing the latest version (e.g. ```latest.json```).
- Hosts the compiled binaries for each supported operating system.

**A client (the actual self-updating application) checks the server on startup to see if a newer version is available. If one is found, it:**

- Downloads the new binary.
- Spawns a short-lived “updater” process to replace itself with the new version.
- Relaunches the application with the updated binary.

## Table of Contents

- Prompt
- Questions & Assumptions
- Repository Structure
 - Prerequisites
- Building
    - Cross-Compilation
- Running the Server
- Running the Client
- How the Self-Update Works
- Caveats & Next Steps

## Prompt

_Write a program that updates itself. Imagine that you have a program you’ve deployed to clients and you will periodically produce new versions. When a new version is produced, we want the deployed programs to be seamlessly replaced by a new version._

_Please write what you consider to be production quality code, whatever that means to you. You may choose any common programming language you like (Go, Rust, or C++ would be good choices). Your program should reasonably be expected to work across common desktop operating systems (Windows, Mac, Linux). If your scheme requires non-trivial server components, please write those as well._

_As in real life, answers to clarifying questions may be elusive. If you have a question, write it down, and guess what you think our answer might be, and proceed as if that were the answer._

## Questions & Assumptions

**How do we determine the version of the client?**

**Assumption:** _We store the current version in a Go constant (e.g., ```const CurrentVersion = "1.0.0"```). In production, you might build the version into the binary using build flags or embed._

**Where and how do we host the server?**

**Assumption:** The project is running via server ```nametag.magnarelli.net```, but the code can be easily changed to run the server & client via ```localhost```.

**How does the client know which binary to download?**

**Assumption:** The client uses ```runtime.GOOS``` to select the correct download URL from a ```latest.json``` file that contains separate URLs for Windows, Darwin (macOS), and Linux.

**Do we pre-build & store the binary on the server, or build the binary locally?**

**Assumption:** As if this was a **production application**, it's best to keep the source code secure - thus why I would reccomend downloading the binary. But in the case for this exercise, the source code is already exposed.

**Is security important for these updates?**

**Assumption:** We’re not yet implementing HTTPS, code signing, checksums, or cryptographic signatures in this sample. In production, these are essential.

**How do we handle the inability to overwrite running executables (particularly on Windows)?**

**Assumption:** We do a small “updater” dance:
- Download the new binary to a temporary file.
- Spawn a child process that deletes the old binary (once the old process exits) and renames/copies the new one.
- The child process restarts the updated application.

## Repository Structure

```
.
├── README.md                        # This file
├── server
│   ├── latest.json                  # JSON metadata
│   └── server.go                    # Minimal server that hosts files
└── client
    ├── main.go                      # Self-updating client code
└── dist
    ├── myapp-windows.exe            # Windows build of the client
    ├── myapp-darwin                 # macOS build of the client 
    └── myapp-linux                  # Linux build of the client
```
## Prerequisites

- Go 1.18+ (or a reasonably recent version).
- Basic familiarity with the terminal/command line.
- Networking access to run the server and have the client retrieve updates.

## Building


**1. Build the Server** ```go build -o server server.go```

You will get a binary called server (or server.exe on Windows).

**2. Build the Self-Updating Client**

**By default, if you just run:**  ```go build -o myapp main.go```

…it will build for your current operating system. However, in order to produce different binaries for distribution (Windows, macOS, Linux), use **cross-compilation**:

## Cross-Compilation

### Windows
```GOOS=windows GOARCH=amd64 go build -o myapp-windows.exe main.go```

### macOS (Darwin)
```GOOS=darwin GOARCH=amd64 go build -o myapp-darwin main.go```

### Linux
```GOOS=linux GOARCH=amd64 go build -o myapp-linux main.go```

_*Adjust amd64 to arm64 or others as needed._
## Running the Server

### Edit latest.json
If running via ```localhost```, ensure it looks something like this:


```
{
  "version": "1.1.0",
  "url_windows": "http://localhost:8080/myapp-windows.exe",
  "url_darwin":  "http://localhost:8080/myapp-darwin",
  "url_linux":   "http://localhost:8080/myapp-linux"
}
```
_Change version to a newer version than the client’s ```CurrentVersion``` if you want the client to update._

### Run the Server
```
./server -port=8080
```
The server will start on http://localhost:8080.

## Running the Client

Run locally (for example, on Linux):
```
./myapp-linux
```
You should see output resembling:

```Hello! I am version 1.0.0
A newer version (1.1.0) is available! Current is 1.0.0
Downloaded new version to /tmp/myapp-update-XYZ
Updater started. Exiting this process to allow the updater to replace the binary.
```

The main process exits. Then the short-lived “updater” runs, replaces ```myapp-linux``` with the new binary, and starts the updated binary. You should see something like:

```Hello! I am version 1.1.0
Continuing normal operation...
Done.
``` 

**On macOS:**

```
chmod +x ./myapp-darwin
./myapp-darwin
```

(Similar output to the Linux example.)

**On Windows:**

```.\myapp-windows.exe```

The logic is essentially the same, though Windows has stricter file-locking behavior. The code works around that with the child “updater” process.

## How the Self-Update Works

#### 1. Check for Update

The client reads ```latest.json``` from the server. If the version differs from its own CurrentVersion, it concludes an update is needed.

#### 2. Download New Binary

The client picks the correct platform-specific URL (e.g., ```url_linux```) and downloads to a temporary file.

#### 3. Spawn the Updater
The client spawns a child process (*itself, but with an ```-update-install``` flag*) to handle the actual file replacement.The main process then exits, freeing its own file lock (important on Windows).

#### 4. Replace the Binary (Child Process)
The child process waits briefly, deletes/renames the old binary, and puts the new one in place. It spawns the new version of the application. It exits.

#### 5. Relaunch
The new version starts up, presumably with an updated ```CurrentVersion```.

## Caveats & Next Steps

### Security
This sample running via localhost uses plain HTTP. In production, use HTTPS. We should also Validate signatures or checksums on downloaded binaries to ensure integrity and authenticity.
### Error Handling
The code is simplified. In a production system, you’d want robust error handling for partial downloads, network timeouts, etc.
### Version Comparison
Our example just checks ```info.Version == CurrentVersion```. Real setups often use semantic version checks.
### Rollback
Might keep a backup of the old binary in case the new version fails to launch.

### Atomicity on Unix
On Unix-like systems, you can often do an ```os.Rename()``` on a running binary. However, on Windows, this is locked. The child-process approach shown here is a cross-platform compromise.

### Additional Platforms
Extend as needed (e.g., ARM builds, 32-bit, etc.).