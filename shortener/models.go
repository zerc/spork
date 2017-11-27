package shortnener

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"os"
)

type ShortURL struct {
	Original string `json:"original_url"`
	Hash     string `json:"-"`
	URL      string `json:"short_url"`
}

func (s *ShortURL) SetURL() error {
	if s.Hash == "" {
		return fmt.Errorf("Hash should be set")
	}

	s.URL = fmt.Sprintf("%s/s/%s", os.Getenv("SPORK_URL"), s.Hash)
	return nil
}

func AllShortURLs(db *sql.DB) ([]*ShortURL, error) {
	rows, err := db.Query("SELECT original, hash FROM urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*ShortURL, 0)

	for rows.Next() {
		a := new(ShortURL)
		if err := rows.Scan(&a.Original, &a.Hash); err != nil {
			return nil, err
		}
		a.SetURL()
		result = append(result, a)
	}

	return result, nil
}

func GetShortURL(db *sql.DB, hash string) (*ShortURL, error) {
	s := ShortURL{Hash: hash}
	err := db.QueryRow("SELECT original FROM urls WHERE hash = $1", hash).Scan(&s.Original)

	if err != nil {
		return nil, err
	}

	s.SetURL()
	return &s, nil
}

func CreateShortURL(db *sql.DB, url string) (*ShortURL, error) {
	hash := md5.New()
	io.WriteString(hash, url)
	hashString := fmt.Sprintf("%x", hash.Sum(nil))[:10]

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE hash = $1)", hashString).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("The record is already exists!")
	} else {
		fmt.Printf("Insert: %s, %s\n", url, hashString)
		_, err := db.Exec("INSERT INTO urls (original, hash) VALUES ($1, $2)", url, hashString)
		if err != nil {
			return nil, err
		}

		s := ShortURL{Original: url, Hash: hashString}
		s.SetURL()
		return &s, nil
	}
}
