package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kr/pretty"
	_ "github.com/lib/pq"
	"googlemaps.github.io/maps"
)

var db *sqlx.DB

func init() {
	var err error

	connStr := `user=your-user dbname=your-db-name sslmode=disable`
	db, err = sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	res, err := fetch()
	if err != nil {
		panic(err)
	}

	insert(res)

	pretty.Print(query(1))
}

func insert(src interface{}) {
	jsonString, err := json.Marshal(src)
	if err != nil {
		panic(err)
	}

	if _, err := db.Exec(`insert into test (opening_hours , updated_at) values ($1, $2) returning *`, jsonString, time.Now()); err != nil {
		panic(err)
	}
}

func query(dayofweek int) *OpeningHours {
	var (
		oph  = OpeningHours{}
		stmt = fmt.Sprintf(`SELECT opening_hours -> 'periods' -> %d FROM test limit 1;`, dayofweek)
	)

	if err := db.QueryRow(stmt).Scan(&oph); err != nil {
		panic(err)
	}

	return &oph
}

func fetch() (maps.PlaceDetailsResult, error) {
	var (
		pid = "ChIJN1t_tDeuEmsRUsoyG83frY4"
		key = "google api key"
	)

	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		panic(err)
	}

	req := maps.PlaceDetailsRequest{
		PlaceID: pid,
		Fields: []maps.PlaceDetailsFieldMask{
			maps.PlaceDetailsFieldMaskOpeningHours,
			maps.PlaceDetailsFieldMaskUTCOffset,
		},
	}

	return c.PlaceDetails(context.Background(), &req)
}

type OpeningHours struct {
	Open  Dtm `json:"open"`
	Close Dtm `json:"close"`
}

type Dtm struct {
	Day  int    `json:"day"`
	Time string `json:"time"`
}

func (o *OpeningHours) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	var oph OpeningHours
	err := json.Unmarshal(source, &oph)
	if err != nil {
		return err
	}
	*o = oph

	return nil
}
