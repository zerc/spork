package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
)

var db *sql.DB

var hostname string = "http://localhost:8000" // TODO: env variable

type ShortURL struct {
	Original string
	Hash     string
}

func (s *ShortURL) All() *[]ShortURL {
	rows, err := db.Query("SELECT original, hash FROM urls")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var result []ShortURL

	for rows.Next() {
		a := ShortURL{}
		if err := rows.Scan(&a.Original, &a.Hash); err != nil {
			log.Fatal(err)
		}
		result = append(result, a)
	}

	return &result
}

func (s *ShortURL) Get() {
	err := db.QueryRow("SELECT original FROM urls WHERE hash = $1", s.Hash).Scan(&s.Original)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *ShortURL) GetShortURL() string {
	return fmt.Sprintf("%s/s/%s", hostname, s.Hash)
}

func (s *ShortURL) Save() error {
	hash := md5.New()
	io.WriteString(hash, s.Original)
	s.Hash = fmt.Sprintf("%x", hash.Sum(nil))[:10]

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE hash = $1)", s.Hash).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		fmt.Println("Can't insert!")
	} else {
		fmt.Printf("Insert: %s, %s\n", s.Original, s.Hash)
		_, err := db.Exec("INSERT INTO urls (original, hash) VALUES ($1, $2)", s.Original, s.Hash)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func init() {
	files, err := ioutil.ReadDir("migrations")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}

	var db_err error
	db, db_err = sql.Open("postgres", "postgres://zero13cool:123@db:5432/shortner?sslmode=disable")
	if db_err != nil {
		log.Fatal(db_err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations/", "postgres", driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}
}

func main() {
	fmt.Println("Server started")
	http.HandleFunc("/api/urls/", ShortURLHandler)
	http.HandleFunc("/s/", RedirectHandler)
	http.HandleFunc("/shortener/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/shortener.html")
	})
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/urls/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		url := r.FormValue("url")

		if url == "" {
			io.WriteString(w, `"url" parameter is missing!`)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			shortURL := ShortURL{Original: url}
			shortURL.Save()
			io.WriteString(w, shortURL.GetShortURL())
			w.WriteHeader(http.StatusOK)
		}

		return
	}

	if r.Method == http.MethodGet {
		rows := ShortURL{}
		b, err := json.Marshal(rows.All())
		if err != nil {
			log.Fatal(err)
		}
		w.Write(b)
		w.Header().Add("Content-Type", "application/json") // TODO: set by default
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Index(r.URL.Path, "/s/") != 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hash := strings.Split(strings.TrimLeft(r.URL.Path, "/s/"), "/")[0]
	fmt.Printf("Got a hash: %s\n", hash)

	if hash == "" {
		io.WriteString(w, "Invalid URL")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := ShortURL{Hash: hash}
	shortURL.Get()

	if shortURL.Original == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(w, r, shortURL.Original, 302)
}
