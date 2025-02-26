// main.go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "runtime"
    "time"
)

// CurrentVersion is the version of this binary.
// In reality, i'd embed this in a more sophisticated way (ldflags, build script, etc.)
const CurrentVersion = "1.0.0"

// We define a struct to match our latest.json
type LatestInfo struct {
    Version    string `json:"version"`
    URLWindows string `json:"url_windows"`
    URLDarwin  string `json:"url_darwin"`
    URLLinux   string `json:"url_linux"`
}

// This flag is used internally to see if we are running as an updater (child) process
var updaterFlag = flag.Bool("update-install", false, "run internal update installer")

func main() {
    flag.Parse()

    // If this is running in "updater" mode, do the file replace dance.
    if *updaterFlag {
        doUpdateInstall()
        return
    }

    // Normal startup
    fmt.Printf("Hello! I am version %s\n", CurrentVersion)

    // For demonstration, let's do a quick check for an update:
    // If you want to run the server and client locally, change the domain to localhost
    checkAndPerformUpdate("https://nametag.magnarelli.net/latest.json")

    // Then do normal operation...
    fmt.Println("Continuing normal operation... (pretend there's more going on here)")
    time.Sleep(5 * time.Second)
    fmt.Println("Done.")
}

// checkAndPerformUpdate checks the server, compares versions, and triggers an update if needed.
func checkAndPerformUpdate(latestURL string) {
    // 1. Fetch the JSON
    info, err := fetchLatestInfo(latestURL)
    if err != nil {
        log.Printf("Failed to fetch latest info: %v", err)
        return
    }

    // 2. Compare versions (simple string compare, or use semver library in real life)
    if info.Version == CurrentVersion {
        log.Println("Already at the latest version.")
        return
    }

    log.Printf("A newer version (%s) is available! Current is %s", info.Version, CurrentVersion)

    // 3. Determine which URL to download
    var downloadURL string
    switch runtime.GOOS {
    case "windows":
        downloadURL = info.URLWindows
    case "darwin":
        downloadURL = info.URLDarwin
    default:
        downloadURL = info.URLLinux
    }

    // 4. Download the new binary to a temp file
    tmpFile, err := downloadFile(downloadURL)
    if err != nil {
        log.Printf("Error downloading file: %v", err)
        return
    }
    defer os.Remove(tmpFile) // Clean up after ourselves

    log.Printf("Downloaded new version to %s", tmpFile)

    // 5. Launch a child process to do the actual install
    // We'll pass the path to the temp file as an argument, so the child process knows what to copy
    selfPath, err := os.Executable()
    if err != nil {
        log.Printf("Could not get executable path: %v", err)
        return
    }

    // On Windows, we might want to quote the path, but let's keep it simple for the example
    cmd := exec.Command(selfPath, "-update-install", tmpFile, selfPath)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 6. Start the updater, then exit this process
    err = cmd.Start()
    if err != nil {
        log.Printf("Could not start updater: %v", err)
        return
    }

    log.Println("Updater started. Exiting this process to allow the updater to replace the binary.")
    os.Exit(0)
}

// fetchLatestInfo fetches and unmarshals the JSON describing the latest version
func fetchLatestInfo(url string) (*LatestInfo, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("non-200 status: %d", resp.StatusCode)
    }

    var info LatestInfo
    if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, err
    }
    return &info, nil
}

// downloadFile downloads the file from `url` to a temporary file.
func downloadFile(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    // Create a temporary file
    tmpf, err := os.CreateTemp("", "myapp-update-")
    if err != nil {
        return "", err
    }
    defer tmpf.Close()

    _, err = io.Copy(tmpf, resp.Body)
    if err != nil {
        return "", err
    }

    // Make it executable
    err = tmpf.Chmod(0755)
    if err != nil {
        // On Windows this might fail, so ignore or log
        log.Printf("Warning: chmod failed: %v", err)
    }

    return tmpf.Name(), nil
}

// doUpdateInstall is the routine that replaces the current binary with the newly downloaded one.
func doUpdateInstall() {
    if len(flag.Args()) < 2 {
        // We expect 2 arguments: [TMPFILE, TARGET_PATH]
        log.Println("Updater usage: -update-install <tmpfile> <targetPath>")
        return
    }
    tmpfile := flag.Args()[0]
    targetPath := flag.Args()[1]

    log.Printf("Updater started with tmpfile=%s, targetPath=%s", tmpfile, targetPath)

    // We need to wait for the main process to exit, but we'll do a small sleep
    // Typically, the main process is about to exit anyway, but let's be safe
    time.Sleep(1 * time.Second)

    // Attempt to remove the old file, then rename the tmpfile to target
    // On Windows, this won't work if the old process is still running, so we rely on it having exited.
    // If it doesn't work, we can try a copy approach.
    err := os.Remove(targetPath)
    if err != nil && !os.IsNotExist(err) {
        log.Printf("Warning: could not remove old file: %v", err)
    }

    // Now rename
    err = os.Rename(tmpfile, targetPath)
    if err != nil {
        log.Printf("Rename failed, fallback to copy approach. Error: %v", err)
        // Fallback to manual copy
        copyErr := copyFileContents(tmpfile, targetPath)
        if copyErr != nil {
            log.Fatalf("Could not copy new file to target: %v", copyErr)
        }
        // remove tmpfile
        os.Remove(tmpfile)
    }

    // Attempt to make sure the new file is executable
    if runtime.GOOS != "windows" {
        os.Chmod(targetPath, 0755)
    }

    // Relaunch the updated binary
    cmd := exec.Command(targetPath)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Start(); err != nil {
        log.Printf("Could not restart the updated binary: %v", err)
    }

    log.Println("Update installed. Exiting updater.")
    os.Exit(0)
}

// copyFileContents is a fallback in case we cannot rename due to OS locking constraints
func copyFileContents(src, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()

    destination, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destination.Close()

    _, err = io.Copy(destination, source)
    return err
}
