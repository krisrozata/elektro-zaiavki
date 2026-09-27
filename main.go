package main

import (
    "log"
    "net/http"

    "elektro-zaiavki/internal/baza"
    "elektro-zaiavki/internal/web"
)

func main() {
    b, err := baza.Otvori("zaiavki.db")
    if err != nil {
        log.Fatal(err)
    }
    defer b.Zatvori()

    server, err := web.NovServer(b, "shabloni")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Сървърът работи на http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", server.Marshruti()))
}