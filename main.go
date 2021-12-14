package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Passenger struct {
	PassengerID int
	FirstName   string
	LastName    string
	MobileNo    string
	EmailAdd    string
}

var db *sql.DB

func validKey(r *http.Request) bool {
	v := r.URL.Query()
	if key, ok := v["key"]; ok {
		if key[0] == "2c78afaf-97da-4816-bbee-9ad239abb296" {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func CreatePassenger(db *sql.DB, p Passenger) {
	query := fmt.Sprintf("INSERT INTO Passenger (FirstName, LastName, MobileNo, EmailAdd) VALUES ('%s','%s','%s','%s')",
		p.FirstName, p.LastName, p.MobileNo, p.EmailAdd)

	_, err := db.Query(query)

	if err != nil {
		panic(err.Error())
	}
}

func GetByID(db *sql.DB, passengerID int) (Passenger, string) {
	query := fmt.Sprintf("SELECT * FROM Passenger where PassengerID = '%d'", passengerID)

	results := db.QueryRow(query)

	var passenger Passenger
	var errMsg string
	var throwAway string

	switch err := results.Scan(&passenger.PassengerID, &throwAway, &passenger.FirstName, &passenger.LastName, &passenger.MobileNo, &passenger.EmailAdd); err {
	case sql.ErrNoRows:
		errMsg = "Account does not exist"
	case nil:
	default:
		panic(err.Error())
	}

	return passenger, errMsg
}

func UpdatePassenger(db *sql.DB, passengerID int, p Passenger) {
	query := fmt.Sprintf("UPDATE Passenger FirstName = '%s', LastName = '%s', MobileNo = '%s', EmailAdd = '%s' WHERE PassengerID = '%d'",
		p.FirstName, p.LastName, p.MobileNo, p.EmailAdd, passengerID)

	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

func GetPassengerDetails(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return

	}
}

func GetPassengerDetailsByID(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	var passengerid int
	var passenger Passenger
	var errMsg string

	passenger, errMsg = GetByID(db, passengerid)
	if errMsg == "Account does not exist" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - No account found"))
	} else {
		json.NewEncoder(w).Encode(passenger)
	}
}

func UpdatePassengerDetails(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	var passengerid int
	reqBody, err := ioutil.ReadAll(r.Body)

	if err == nil {
		var passenger Passenger
		json.Unmarshal([]byte(reqBody), &passenger)

		if passenger.FirstName == "" || passenger.LastName == "" || passenger.MobileNo == "" || passenger.EmailAdd == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Please supply all neccessary passenger information "))
		} else {
			UpdatePassenger(db, passengerid, passenger)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - Account details updated"))
		}
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("422 - Please supply passenger information in JSON format"))
	}
}

func main() {

	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/passengers")

	// handle error
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

}
