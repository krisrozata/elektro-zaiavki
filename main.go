package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "slices"
    "strconv"
    "time"

    _ "modernc.org/sqlite"
)

type Stranica struct {
    Zaglavie string
    Opisanie string
}

type Kontakti struct {
    Telefon      string
    Adres        string
    RabotnoVreme string
}

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

type StranicaSpisak struct {
    Broi    int
    Zaiavki []Zaiavka
    Statusi []string
}

var statusi = []string{"Нова", "В процес", "Завършена", "Отказана"}

var shabloni = template.Must(template.ParseGlob("shabloni/*.html"))

var db *sql.DB

func otvoriBaza() (*sql.DB, error) {
    baza, err := sql.Open("sqlite", "zaiavki.db")
    if err != nil {
        return nil, err
    }

    _, err = baza.Exec(`
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
        baza.Close()
        return nil, err
    }

    return baza, nil
}

func nachalo(w http.ResponseWriter, r *http.Request) {
    danni := Stranica{
        Zaglavie: "Електро заявки",
        Opisanie: "Ел. ремонти и монтаж в Пловдив",
    }
    pokazhi(w, "nachalo.html", danni)
}

func kontakti(w http.ResponseWriter, r *http.Request) {
    danni := Kontakti{
        Telefon:      "0888 123 456",
        Adres:        "Пловдив",
        RabotnoVreme: "Пон–Пет, 8:00–18:00",
    }
    pokazhi(w, "kontakti.html", danni)
}

func formaZaiavka(w http.ResponseWriter, r *http.Request) {
    pokazhi(w, "zaiavka.html", nil)
}

func priemiZaiavka(w http.ResponseWriter, r *http.Request) {
    ime := r.FormValue("ime")
    telefon := r.FormValue("telefon")

    if ime == "" || telefon == "" {
        http.Error(w, "Името и телефонът са задължителни", http.StatusBadRequest)
        return
    }

    _, err := db.Exec(`
        INSERT INTO zaiavki (ime, telefon, adres, vid, opisanie, suzdadena)
        VALUES (?, ?, ?, ?, ?, ?)`,
        ime,
        telefon,
        r.FormValue("adres"),
        r.FormValue("vid"),
        r.FormValue("opisanie"),
        time.Now().Unix(),
    )
    if err != nil {
        log.Println("Грешка при запис:", err)
        http.Error(w, "Заявката не можа да бъде записана", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/zaiavki", http.StatusSeeOther)
}

func spisakZaiavki(w http.ResponseWriter, r *http.Request) {
    redove, err := db.Query(`
        SELECT id, ime, telefon, adres, vid, opisanie, status, suzdadena
        FROM zaiavki
        ORDER BY id DESC`)
    if err != nil {
        log.Println("Грешка при четене:", err)
        http.Error(w, "Заявките не можаха да бъдат заредени", http.StatusInternalServerError)
        return
    }
    defer redove.Close()

    var zaiavki []Zaiavka
    for redove.Next() {
        var z Zaiavka
        var vreme int64

        err := redove.Scan(&z.ID, &z.Ime, &z.Telefon, &z.Adres, &z.Vid, &z.Opisanie, &z.Status, &vreme)
        if err != nil {
            log.Println("Грешка при четене на ред:", err)
            http.Error(w, "Заявките не можаха да бъдат заредени", http.StatusInternalServerError)
            return
        }

        z.Suzdadena = time.Unix(vreme, 0)
        zaiavki = append(zaiavki, z)
    }

    if err := redove.Err(); err != nil {
        log.Println("Грешка след четене:", err)
        http.Error(w, "Заявките не можаха да бъдат заредени", http.StatusInternalServerError)
        return
    }

    var broi int
    err = db.QueryRow(`SELECT COUNT(*) FROM zaiavki`).Scan(&broi)
    if err != nil {
        log.Println("Грешка при броене:", err)
        http.Error(w, "Заявките не можаха да бъдат заредени", http.StatusInternalServerError)
        return
    }

    danni := StranicaSpisak{
        Broi:    broi,
        Zaiavki: zaiavki,
        Statusi: statusi,
    }
    pokazhi(w, "spisak.html", danni)
}

func smeniStatus(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "Невалиден номер на заявка", http.StatusBadRequest)
        return
    }

    novStatus := r.FormValue("status")
    if !slices.Contains(statusi, novStatus) {
        http.Error(w, "Невалиден статус", http.StatusBadRequest)
        return
    }

    rezultat, err := db.Exec(`UPDATE zaiavki SET status = ? WHERE id = ?`, novStatus, id)
    if err != nil {
        log.Println("Грешка при смяна на статус:", err)
        http.Error(w, "Статусът не можа да бъде сменен", http.StatusInternalServerError)
        return
    }

    promeneni, err := rezultat.RowsAffected()
    if err == nil && promeneni == 0 {
        http.Error(w, "Няма такава заявка", http.StatusNotFound)
        return
    }

    http.Redirect(w, r, "/zaiavki", http.StatusSeeOther)
}

func pokazhi(w http.ResponseWriter, ime string, danni any) {
    err := shabloni.ExecuteTemplate(w, ime, danni)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    var err error
    db, err = otvoriBaza()
    if err != nil {
        log.Fatal("Базата не може да се отвори: ", err)
    }
    defer db.Close()

    http.HandleFunc("/", nachalo)
    http.HandleFunc("/kontakti", kontakti)
    http.HandleFunc("GET /zaiavka", formaZaiavka)
    http.HandleFunc("POST /zaiavka", priemiZaiavka)
    http.HandleFunc("GET /zaiavki", spisakZaiavki)
    http.HandleFunc("POST /zaiavki/{id}/status", smeniStatus)

    fmt.Println("Сървърът работи на http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}