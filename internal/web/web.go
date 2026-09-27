package web

import (
    "crypto/sha256"
    "crypto/subtle"
    "errors"
    "html/template"
    "log"
    "net/http"
    "strconv"
    "time"

    "elektro-zaiavki/internal/baza"
)
type Server struct {
    baza        *baza.Baza
    shabloni    *template.Template
    adminIme    string
    adminParola string
}

type stranica struct {
    Zaglavie string
    Opisanie string
}

type kontaktiDanni struct {
    Telefon      string
    Adres        string
    RabotnoVreme string
}

type stranicaSpisak struct {
    Broi    int
    Zaiavki []baza.Zaiavka
    Statusi []string
}

func NovServer(b *baza.Baza, papkaShabloni, adminIme, adminParola string) (*Server, error) {
    t, err := template.ParseGlob(papkaShabloni + "/*.html")
    if err != nil {
        return nil, err
    }
    return &Server{
        baza:        b,
        shabloni:    t,
        adminIme:    adminIme,
        adminParola: adminParola,
    }, nil
}

func (s *Server) Marshruti() *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("/", s.nachalo)
    mux.HandleFunc("/kontakti", s.kontakti)
    mux.HandleFunc("GET /zaiavka", s.formaZaiavka)
    mux.HandleFunc("POST /zaiavka", s.priemiZaiavka)
    mux.HandleFunc("GET /blagodarim", s.blagodarim)

    mux.HandleFunc("GET /zaiavki", s.samoAdmin(s.spisakZaiavki))
    mux.HandleFunc("POST /zaiavki/{id}/status", s.samoAdmin(s.smeniStatus))
    return mux
}
func (s *Server) nachalo(w http.ResponseWriter, r *http.Request) {
    s.pokazhi(w, "nachalo.html", stranica{
        Zaglavie: "Електро заявки",
        Opisanie: "Ел. ремонти и монтаж в Пловдив",
    })
}

func (s *Server) kontakti(w http.ResponseWriter, r *http.Request) {
    s.pokazhi(w, "kontakti.html", kontaktiDanni{
        Telefon:      "0888 123 456",
        Adres:        "Пловдив",
        RabotnoVreme: "Пон–Пет, 8:00–18:00",
    })
}

func (s *Server) formaZaiavka(w http.ResponseWriter, r *http.Request) {
    s.pokazhi(w, "zaiavka.html", nil)
}

func (s *Server) priemiZaiavka(w http.ResponseWriter, r *http.Request) {
    z := baza.Zaiavka{
        Ime:       r.FormValue("ime"),
        Telefon:   r.FormValue("telefon"),
        Adres:     r.FormValue("adres"),
        Vid:       r.FormValue("vid"),
        Opisanie:  r.FormValue("opisanie"),
        Suzdadena: time.Now(),
    }

    if z.Ime == "" || z.Telefon == "" {
        http.Error(w, "Името и телефонът са задължителни", http.StatusBadRequest)
        return
    }

    if err := s.baza.Dobavi(z); err != nil {
        log.Println(err)
        http.Error(w, "Заявката не можа да бъде записана", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/blagodarim", http.StatusSeeOther)
}

func (s *Server) spisakZaiavki(w http.ResponseWriter, r *http.Request) {
    zaiavki, err := s.baza.Vsichki()
    if err != nil {
        log.Println(err)
        http.Error(w, "Заявките не можаха да бъдат заредени", http.StatusInternalServerError)
        return
    }

    broi, err := s.baza.Broi()
    if err != nil {
        log.Println(err)
        http.Error(w, "Заявките не можаха да бъдат заредени", http.StatusInternalServerError)
        return
    }

    s.pokazhi(w, "spisak.html", stranicaSpisak{
        Broi:    broi,
        Zaiavki: zaiavki,
        Statusi: baza.Statusi,
    })
}

func (s *Server) smeniStatus(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "Невалиден номер на заявка", http.StatusBadRequest)
        return
    }

    novStatus := r.FormValue("status")
    if !baza.ValidenStatus(novStatus) {
        http.Error(w, "Невалиден статус", http.StatusBadRequest)
        return
    }

    err = s.baza.SmeniStatus(id, novStatus)
    if errors.Is(err, baza.ErrNiamaZaiavka) {
        http.Error(w, "Няма такава заявка", http.StatusNotFound)
        return
    }
    if err != nil {
        log.Println(err)
        http.Error(w, "Статусът не можа да бъде сменен", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/zaiavki", http.StatusSeeOther)
}

func (s *Server) pokazhi(w http.ResponseWriter, ime string, danni any) {
    if err := s.shabloni.ExecuteTemplate(w, ime, danni); err != nil {
        log.Println(err)
        http.Error(w, "Грешка при показване на страницата", http.StatusInternalServerError)
    }
}
func (s *Server) blagodarim(w http.ResponseWriter, r *http.Request) {
    s.pokazhi(w, "blagodarim.html", nil)
}

func (s *Server) samoAdmin(sledvasht http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ime, parola, ok := r.BasicAuth()

        if !ok || !ednakvi(ime, s.adminIme) || !ednakvi(parola, s.adminParola) {
            w.Header().Set("WWW-Authenticate", `Basic realm="Заявки", charset="UTF-8"`)
            http.Error(w, "Нужен е вход", http.StatusUnauthorized)
            return
        }

        sledvasht(w, r)
    }
}

func ednakvi(a, b string) bool {
    hashA := sha256.Sum256([]byte(a))
    hashB := sha256.Sum256([]byte(b))
    return subtle.ConstantTimeCompare(hashA[:], hashB[:]) == 1
}