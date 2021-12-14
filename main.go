package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
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

//==================== Database functions ====================
func UniquenessValidation(db *sql.DB, username string, emailAddress string) string {
	query := fmt.Sprintf("SELECT * FROM Passenger where Username = '%s' or emailAddress = '%s'", username, emailAddress)

	results := db.QueryRow(query)

	errMsg := ""
	var userName string
	var email string

	var throwAway int
	var throwAway2 string

	switch err := results.Scan(&throwAway, &userName, &throwAway2, &throwAway2, &throwAway2, &throwAway2, &email); err {
	case sql.ErrNoRows:
	case nil:
		if userName == username {
			errMsg += "Username already in use. "
		}
		if email == emailAddress {
			errMsg += "Email Address already in use."
		}
	default:
		panic(err.Error())
	}

	return errMsg
}

func CreatePassenger(db *sql.DB, p Passenger) {
	query := fmt.Sprintf("INSERT INTO Passenger (Username, `Password`, FirstName, LastName, MobileNo, EmailAddress) VALUES ('%s','%s','%s','%s','%s','%s')",
		p.FirstName, p.LastName, p.MobileNo, p.EmailAdd)

	_, err := db.Query(query)

	if err != nil {
		panic(err.Error())
	}
}

func Login(db *sql.DB, username string, password string) (Passenger, string) {
	query := fmt.Sprintf("SELECT * FROM Passenger where Username = '%s' and `Password` = '%s'", username, password)

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
	query := fmt.Sprintf("UPDATE Passenger FirstName = '%s', LastName = '%s', MobileNo = '%s', EmailAddress = '%s' WHERE PassengerID = '%d'",
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

	params := mux.Vars(r)

	var passenger Passenger
	var errMsg string

	passenger, errMsg = Login(db, params["username"], params["password"])
	if errMsg == "Account does not exist" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - No account found"))
	} else {
		json.NewEncoder(w).Encode(passenger)
	}
}

func GetPassengerDetailsByID(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	params := mux.Vars(r)
	var passengerid int
	fmt.Sscan(params["passengerid"], &passengerid)

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

	params := mux.Vars(r)
	var passengerid int
	fmt.Sscan(params["passengerid"], &passengerid)

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

//==================== Main ====================
func main() {

	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/passengers")

	// handle error
	if err != nil {
		panic(err.Error())
	}

	// defer the close till after the main function has finished executing
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/passengers?username={username}&password={password}", GetPassengerDetails).Methods("GET")
	router.HandleFunc("/api/v1/passengers/{passengerid}", GetPassengerDetailsByID).Methods("GET")
	router.HandleFunc("/api/v1/passengers/{passengerid}", UpdatePassengerDetails).Methods("PUT")

	fmt.Println("Passenger Service operating on port 5000")
	log.Fatal(http.ListenAndServe(":5000", router))
}
