package main

import (
    "log"
    "net/http"
    "os"

    "elektro-zaiavki/internal/baza"
    "elektro-zaiavki/internal/web"
)

func main() {
    adminIme := os.Getenv("ADMIN_IME")
    if adminIme == "" {
        adminIme = "admin"
    }

    adminParola := os.Getenv("ADMIN_PAROLA")
    if adminParola == "" {
        log.Fatal("Липсва ADMIN_PAROLA. Задай парола, преди да стартираш сървъра.")
    }

    b, err := baza.Otvori("zaiavki.db")
    if err != nil {
        log.Fatal(err)
    }
    defer b.Zatvori()

    server, err := web.NovServer(b, "shabloni", adminIme, adminParola)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Сървърът работи на http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", server.Marshruti()))
}