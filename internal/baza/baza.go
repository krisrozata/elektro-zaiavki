package baza

import (
    "database/sql"
    "errors"
    "fmt"
    "slices"
    "time"

    _ "modernc.org/sqlite"
)

var ErrNiamaZaiavka = errors.New("няма такава заявка")

var Statusi = []string{"Нова", "В процес", "Завършена", "Отказана"}

type Zaiavka struct {
    ID        int
    Ime       string
    Telefon   string
    Adres     string
    Vid       string
    Opisanie  string
    Status    string
    Suzdadena time.Time
}

type Baza struct {
    db *sql.DB
}

func Otvori(pat string) (*Baza, error) {
    db, err := sql.Open("sqlite", pat)
    if err != nil {
        return nil, fmt.Errorf("отваряне на базата: %w", err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS zaiavki (
            id        INTEGER PRIMARY KEY AUTOINCREMENT,
            ime       TEXT NOT NULL,
            telefon   TEXT NOT NULL,
            adres     TEXT NOT NULL DEFAULT '',
            vid       TEXT NOT NULL DEFAULT '',
            opisanie  TEXT NOT NULL DEFAULT '',
            status    TEXT NOT NULL DEFAULT 'Нова',
            suzdadena INTEGER NOT NULL
        )`)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("създаване на таблицата: %w", err)
    }

    return &Baza{db: db}, nil
}

func (b *Baza) Zatvori() error {
    return b.db.Close()
}

func ValidenStatus(s string) bool {
    return slices.Contains(Statusi, s)
}

func (b *Baza) Dobavi(z Zaiavka) error {
    _, err := b.db.Exec(`
        INSERT INTO zaiavki (ime, telefon, adres, vid, opisanie, suzdadena)
        VALUES (?, ?, ?, ?, ?, ?)`,
        z.Ime, z.Telefon, z.Adres, z.Vid, z.Opisanie, z.Suzdadena.Unix(),
    )
    if err != nil {
        return fmt.Errorf("запис на заявка: %w", err)
    }
    return nil
}

func (b *Baza) Vsichki() ([]Zaiavka, error) {
    redove, err := b.db.Query(`
        SELECT id, ime, telefon, adres, vid, opisanie, status, suzdadena
        FROM zaiavki
        ORDER BY id DESC`)
    if err != nil {
        return nil, fmt.Errorf("четене на заявки: %w", err)
    }
    defer redove.Close()

    var zaiavki []Zaiavka
    for redove.Next() {
        var z Zaiavka
        var vreme int64

        err := redove.Scan(&z.ID, &z.Ime, &z.Telefon, &z.Adres, &z.Vid, &z.Opisanie, &z.Status, &vreme)
        if err != nil {
            return nil, fmt.Errorf("четене на ред: %w", err)
        }

        z.Suzdadena = time.Unix(vreme, 0)
        zaiavki = append(zaiavki, z)
    }

    if err := redove.Err(); err != nil {
        return nil, fmt.Errorf("след четене: %w", err)
    }

    return zaiavki, nil
}

func (b *Baza) Broi() (int, error) {
    var broi int
    err := b.db.QueryRow(`SELECT COUNT(*) FROM zaiavki`).Scan(&broi)
    if err != nil {
        return 0, fmt.Errorf("броене на заявки: %w", err)
    }
    return broi, nil
}

func (b *Baza) SmeniStatus(id int, status string) error {
    rezultat, err := b.db.Exec(`UPDATE zaiavki SET status = ? WHERE id = ?`, status, id)
    if err != nil {
        return fmt.Errorf("смяна на статус: %w", err)
    }

    promeneni, err := rezultat.RowsAffected()
    if err != nil {
        return fmt.Errorf("проверка на промяната: %w", err)
    }
    if promeneni == 0 {
        return ErrNiamaZaiavka
    }

    return nil
}