package main

import (
    "fmt"
    "net/http"
)

func nachalo(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Здравей от Go сървъра!")
}

func kontakti(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Контакти")
    fmt.Fprintln(w, "Телефон: 0888 123 456")
    fmt.Fprintln(w, "Адрес: Пловдив")
}

func main() {
    http.HandleFunc("/", nachalo)
    http.HandleFunc("/kontakti", kontakti)

    fmt.Println("Сървърът работи на http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Грешка:", err)
    }
}