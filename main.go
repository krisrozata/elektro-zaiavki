package main

import (
    "fmt"
    "net/http"
)

func nachalo(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Здравей от Go сървъра!")
}

func main() {
    http.HandleFunc("/", nachalo)

    fmt.Println("Сървърът работи на http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Грешка:", err)
    }
}
