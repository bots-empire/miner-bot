package model

import (
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/pkg/errors"
)

const (
	AvailableSymbolInHash = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	HashKeyLength         = 10

	botLink = "%s?start=%s"
)

type ReferralLinkInfo struct {
	HashKey    string
	ReferralID int64
	Source     string
}

// EncodeLink generates a link and saves user data to the database
func EncodeLink(database *sql.DB, botLink string, link *ReferralLinkInfo) (string, error) {
	link.HashKey = getHash()

	if err := saveLinkInDataBase(database, link); err != nil {
		return "", errors.Wrap(err, "save link in database")
	}

	return fmt.Sprintf(botLink, botLink, link.HashKey), nil
}

func getHash() string {
	var key string

	rs := []rune(AvailableSymbolInHash)
	lenOfArray := len(rs)

	for i := 0; i < HashKeyLength; i++ {
		key += string(rs[rand.Intn(lenOfArray)])
	}
	return key
}

func saveLinkInDataBase(dataBase *sql.DB, link *ReferralLinkInfo) error {
	_, err := dataBase.Exec("INSERT INTO links VALUES (?, ?, ?)",
		link.HashKey,
		link.ReferralID,
		link.Source)
	if err != nil {
		return errors.Wrap(err, "make exec in database")
	}

	return nil
}

// DecodeLink using a hash from the database returns information about the link
func DecodeLink(dataBase *sql.DB, hashKey string) (*ReferralLinkInfo, error) {
	rows, err := dataBase.Query("SELECT * FROM links WHERE hash = ?",
		hashKey)
	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}
	defer rows.Close()

	linkInfo, err := scanLinkFromRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "scan link info")
	}

	return linkInfo, nil
}

func scanLinkFromRows(rows *sql.Rows) (*ReferralLinkInfo, error) {
	for rows.Next() {
		link := &ReferralLinkInfo{}

		err := rows.Scan(
			&link.HashKey,
			&link.ReferralID,
			&link.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed scan row")
		}

		return link, nil
	}

	return nil, nil
}
