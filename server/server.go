// server.go
package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
)

func main() {
    port := flag.Int("port", 8080, "port to serve on")
    flag.Parse()

    // current working directory
    cwd, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }

    // Serve files in current directory (where we presumably have latest.json and the binaries)
    http.Handle("/", http.FileServer(http.Dir(cwd)))

    addr := fmt.Sprintf(":%d", *port)
    log.Printf("Serving on http://localhost%s ...", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatal(err)
    }
}
