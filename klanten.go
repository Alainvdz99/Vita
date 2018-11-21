package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"log"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

// Gegevens van de klanten die getoond, ingevoerd, gewijzigd of verwijderd kunnen worden
type Klant struct {
	Knr, Hnr, Ink                    int64
	Nm, Vnm, Pc, Hnrt, Gsl, Blg, Rhf string
	Gbd, Krg, Opl, Opm               string
	Brf                              float64
}

type Bestelling struct {
	Bnr, Dlt, Knr, Vkp     int
	Sts, Bsd               string
	Bdr                    float32
}

type Module struct {
	Mnm, Oms	string
	Spr 		float32
}

// Verbinding maken met de database
func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "vitaintellectdb"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

// De routing die afgeleid wordt naar de template
var tmpl = template.Must(template.ParseGlob("form/*"))

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

//Sessie instellen
func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

// Inlogpagina instellen
func loginHandler(response http.ResponseWriter, request *http.Request) {
	db := dbConn()

	//Gebruikersnaam en wachtwoord van de webpagina lezen
	username := request.FormValue("username")
	pass := request.FormValue("password")

	var databaseUsername string
	var databasePassword string

	//Gebruiker in de database selecteren
	err := db.QueryRow("SELECT voorletters, datum_in_dienst FROM medewerker WHERE voorletters=? AND datum_in_dienst=?", username, pass).Scan(&databaseUsername, &databasePassword)

	if err != nil {
		http.Redirect(response, request, "/login", 301)
		return
	}

	redirectTarget := "/"
	if username != "" && pass != "" {
		// .. check credentials ..
		setSession(username, response)
		redirectTarget = "/internal"
	}
	http.Redirect(response, request, redirectTarget, 302)
}

// logout handler
func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

// index page
const indexPage = `
<h1>Inloggen</h1>
<form method="post" action="/login">
   <label for="username">Gebruikersnaam</label>
   <input type="text" id="username" name="username" placeholder="gebruikersnaam">
   <label for="password">Wachtwoord</label>
   <input type="password" id="password" name="password" placeholder="wachtwoord">
   <button type="submit">Inloggen</button>
</form>
`
func indexPageHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, indexPage)
}


// internal page
const internalPage = `<html><body>
<h1>Venditor: Vita Intellectus</h1>

<hr>
<p>Hallo: %s </p>
<p>
<a href="/index">Mijn klanten</a> |
<a href="/bestelling-index">Mijn bestellingen</a> |
<a href="/modules">Modules</a>
</p>
<form method="post" action="/logout">
    <button type="submit">Uitloggen</button>
</form>
</body></html>
`

func internalPageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)

	if userName != "" {
		fmt.Fprintf(response, internalPage, userName)

	} else {
		http.Redirect(response, request, "/", 302)
	}
}

/**
	KLANTEN
 */

//Index van alle klanten
func Index(w http.ResponseWriter, request *http.Request) {
	db := dbConn()
	userName := getUserName(request)
	selDB, err := db.Query("SELECT k.* FROM klant AS k INNER JOIN bestelling AS b ON b.klantnummer = k.klantnummer WHERE  b.verkoper = (SELECT m.medewerkernummer FROM medewerker AS m WHERE m.voorletters = ?)", userName) // Selecteren en ordenen van de gegevens van de klanten
	if err != nil {
		panic(err.Error())
	}
	klnt := Klant{}
	res := []Klant{}
	for selDB.Next() {
		var klantnummer, huisnummer, inkomen sql.NullInt64
		var naam, voornaam, postcode, huisnummer_toevoeging, geslacht, bloedgroep, rhesusfactor sql.NullString
		var geboortedatum, kredietregistratie, opleiding, opmerkingen sql.NullString
		var beroepsrisicofactor sql.NullFloat64
		err = selDB.Scan(&klantnummer, &naam, &voornaam, &postcode, &huisnummer, &huisnummer_toevoeging, &geboortedatum, &geslacht, &bloedgroep, &rhesusfactor, &beroepsrisicofactor, &inkomen, &kredietregistratie, &opleiding, &opmerkingen)
		if err != nil {
			panic(err.Error())
		}
		klnt.Knr = klantnummer.Int64
		klnt.Nm = naam.String
		klnt.Vnm = voornaam.String
		klnt.Pc = postcode.String
		klnt.Hnr = huisnummer.Int64
		klnt.Hnrt = huisnummer_toevoeging.String
		klnt.Gbd = geboortedatum.String
		klnt.Gsl = geslacht.String
		klnt.Blg = bloedgroep.String
		klnt.Rhf = rhesusfactor.String
		klnt.Brf = beroepsrisicofactor.Float64
		klnt.Ink = inkomen.Int64
		klnt.Krg = kredietregistratie.String
		klnt.Opl = opleiding.String
		klnt.Opm = opmerkingen.String
		res = append(res, klnt)
	}
	//tmpl.ExecuteTemplate(w, "Index", nil)
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
}

// De gegevens van één klant tonen
func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nKnr := r.URL.Query().Get("klantnummer")
	selDB, err := db.Query("SELECT * FROM klant WHERE klantnummer=?", nKnr)
	if err != nil {
		panic(err.Error())
	}
	klnt := Klant{}
	for selDB.Next() {
		var klantnummer, huisnummer, inkomen sql.NullInt64
		var naam, voornaam, postcode, huisnummer_toevoeging, geslacht, bloedgroep, rhesusfactor sql.NullString
		var geboortedatum, kredietregistratie, opleiding, opmerkingen sql.NullString
		var beroepsrisicofactor sql.NullFloat64
		err = selDB.Scan(&klantnummer, &naam, &voornaam, &postcode, &huisnummer, &huisnummer_toevoeging, &geboortedatum, &geslacht, &bloedgroep, &rhesusfactor, &beroepsrisicofactor, &inkomen, &kredietregistratie, &opleiding, &opmerkingen)
		if err != nil {
			panic(err.Error())
		}
		klnt.Knr = klantnummer.Int64
		klnt.Nm = naam.String
		klnt.Vnm = voornaam.String
		klnt.Pc = postcode.String
		klnt.Hnr = huisnummer.Int64
		klnt.Hnrt = huisnummer_toevoeging.String
		klnt.Gbd = geboortedatum.String
		klnt.Gsl = geslacht.String
		klnt.Blg = bloedgroep.String
		klnt.Rhf = rhesusfactor.String
		klnt.Brf = beroepsrisicofactor.Float64
		klnt.Ink = inkomen.Int64
		klnt.Krg = kredietregistratie.String
		klnt.Opl = opleiding.String
		klnt.Opm = opmerkingen.String
	}
	tmpl.ExecuteTemplate(w, "Show", klnt)
	defer db.Close()
}

// Nieuwe klant toevoegen
func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

//  Gegevens klant toevoegen aan de database
func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		klantnummer := r.FormValue("klantnummer")
		voornaam := r.FormValue("voornaam")
		naam := r.FormValue("naam")
		postcode := r.FormValue("postcode")
		huisnummer := r.FormValue("huisnummer")
		huisnummer_toevoeging := r.FormValue("huisnummer_toevoeging")
		geboortedatum := r.FormValue("geboortedatum")
		geslacht := r.FormValue("geslacht")
		bloedgroep := r.FormValue("bloedgroep")
		rhesusfactor := r.FormValue("rhesusfactor")
		beroepsrisicofactor := r.FormValue("beroepsrisicofactor")
		inkomen := r.FormValue("inkomen")
		kredietregistratie := r.FormValue("kredietregistratie")
		opleiding := r.FormValue("opleiding")
		opmerkingen := r.FormValue("opmerkingen")
		insForm, err := db.Prepare("INSERT INTO klant(klantnummer, naam, voornaam, postcode, huisnummer, huisnummer_toevoeging, geboortedatum, geslacht, bloedgroep, rhesusfactor, beroepsrisicofactor, inkomen, kredietregistratie, opleiding, opmerkingen) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(klantnummer, naam, voornaam, postcode, huisnummer, huisnummer_toevoeging, geboortedatum, geslacht, bloedgroep, rhesusfactor, beroepsrisicofactor, inkomen, kredietregistratie, opleiding, opmerkingen)
		log.Println("Toegevoegd: Voornaam: " + voornaam + " | Achternaam: " + naam + " | Geboortedatum: " + geboortedatum)
	}
	defer db.Close()
	http.Redirect(w, r, "/index", 301)
}

// Klant verwijderen uit de database
func Delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	std := r.URL.Query().Get("klantnummer")
	delForm, err := db.Prepare("DELETE FROM klant WHERE klantnummer=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(std)
	log.Println("DELETE")
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

/**
	BESTELLINGEN!!!
 */

//Index van alle bestellingen
func bestellingIndex(w http.ResponseWriter, request *http.Request) {
	db := dbConn()
	userName := getUserName(request)
	selDB, err := db.Query("SELECT b.* FROM bestelling AS b WHERE b.verkoper =  (SELECT m.medewerkernummer FROM medewerker AS m WHERE m.voorletters = ? )", userName) // Selecteren en ordenen van de gegevens van de klanten
	if err != nil {
		panic(err.Error())
	}
	bstl := Bestelling{}
	res := []Bestelling{}
	for selDB.Next() {
		var bestelnummer, afbetaling_doorlooptijd, klantnummer, verkoper int
		var status, besteldatum string
		var afbetaling_bestelbedrag float32
		err = selDB.Scan(&bestelnummer, &status, &besteldatum, &afbetaling_doorlooptijd, &afbetaling_bestelbedrag, &klantnummer, &verkoper)
		if err != nil {
			panic(err.Error())
		}
		bstl.Knr = klantnummer
		bstl.Bnr = bestelnummer
		bstl.Dlt = afbetaling_doorlooptijd
		bstl.Vkp = verkoper
		bstl.Sts = status
		bstl.Bsd = besteldatum
		bstl.Bdr = afbetaling_bestelbedrag
		res = append(res, bstl)
	}
	tmpl.ExecuteTemplate(w, "BestellingIndex", res)
	defer db.Close()
}

// De gegevens van één bestelling tonen
func bestellingShow(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nBnr := r.URL.Query().Get("bestelnummer")
	selDB, err := db.Query("SELECT * FROM bestelling WHERE bestelnummer=?", nBnr)
	if err != nil {
		panic(err.Error())
	}
	bstl := Bestelling{}
	for selDB.Next() {
		var bestelnummer, afbetaling_doorlooptijd, klantnummer, verkoper int
		var status, besteldatum string
		var afbetaling_bestelbedrag float32
		err = selDB.Scan(&klantnummer, &bestelnummer, &afbetaling_doorlooptijd, &verkoper, &status, &besteldatum, &afbetaling_bestelbedrag,)
		if err != nil {
			panic(err.Error())
		}
		bstl.Knr = klantnummer
		bstl.Bnr = bestelnummer
		bstl.Dlt = afbetaling_doorlooptijd
		bstl.Vkp = verkoper
		bstl.Sts = status
		bstl.Bsd = besteldatum
		bstl.Bdr = afbetaling_bestelbedrag
	}
	tmpl.ExecuteTemplate(w, "BestellingShow", bstl)
	defer db.Close()
}

// Nieuwe bestelling toevoegen
func bestellingNew(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "BestellingNew", nil)
}

//  Gegevens bestelling toevoegen aan de database
func bestellingInsert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		bestelnummer := r.FormValue("bestelnummer")
		status := r.FormValue("status")
		besteldatum := r.FormValue("naam")
		afbetaling_doorlooptijd := r.FormValue("afbetaling_doorlooptijd")
		afbetaling_bestelbedrag := r.FormValue("afbetaling_bestelbedrag")
		klantnummer := r.FormValue("klantnummer")
		verkoper := r.FormValue("verkoper")
		insForm, err := db.Prepare("INSERT INTO bestelling(bestelnummer, status, besteldatum, afbetaling_doorlooptijd, afbetaling_bestelbedrag, klantnummer, verkoper) VALUES(?,?,?,?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(bestelnummer, status, besteldatum, afbetaling_doorlooptijd, afbetaling_bestelbedrag, klantnummer, verkoper)
	}
	defer db.Close()
	http.Redirect(w, r, "/bestelling-index", 301)
}

// bestelling verwijderen uit de database
func bestellingDelete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	bstl := r.URL.Query().Get("bestelnummer")
	delForm, err := db.Prepare("DELETE FROM bestelling WHERE bestelnummer=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(bstl)
	log.Println("DELETE")
	defer db.Close()
	http.Redirect(w, r, "/bestelling-index", 301)
}

/**
		MODULES!!!
*/

// Weergeeft de modules
func Modules(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM module ORDER BY stukprijs DESC")
	if err != nil {
		panic(err.Error())
	}
	mdl := Module{}
	res := []Module{}
	for selDB.Next() {
		var modulenaam, omschrijving string
		var stukprijs float32
		err = selDB.Scan(&modulenaam, &omschrijving, &stukprijs)
		if err != nil {
			panic(err.Error())
		}
		mdl.Mnm = modulenaam
		mdl.Oms = omschrijving
		mdl.Spr = stukprijs

		res = append(res, mdl)
	}
	tmpl.ExecuteTemplate(w, "Modules", res)
	defer db.Close()
}


// De gegevens tonen
var router = mux.NewRouter()

func main() {
	log.Println("Server started on: http://localhost:8080")
	// login
	router.HandleFunc("/", indexPageHandler)
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")
	router.HandleFunc("/internal", internalPageHandler)
	http.Handle("/", router)
	// klanten
	http.HandleFunc("/index", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/new", New)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/delete", Delete)
	// bestellingen
	http.HandleFunc("/bestelling-index", bestellingIndex)
	http.HandleFunc("/bestelling-show", bestellingShow)
	http.HandleFunc("/bestelling-new", bestellingNew)
	http.HandleFunc("/bestelling-insert", bestellingInsert)
	http.HandleFunc("/bestelling-delete", bestellingDelete)
	// Modules schema
	http.HandleFunc("/modules", Modules)
	// Start server
	http.ListenAndServe(":8080", nil)
}
