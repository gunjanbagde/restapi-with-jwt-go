package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"dev.azure.com/tanla/tanlamvp/_git/common.git/helper/log"
	jwt "github.com/dgrijalva/jwt-go"
)

//jwtKey is used in jwt token to sign the jwt token
var jwtKey = []byte("secret_key")

//storing credentials in map
var users = map[string]string{
	"gunjan806": "Qwerty@123",
	"user2":     "password2",
}

//struct to capture the login request body
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//to create payload inside the jwt
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func Login(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var credentials Credentials
	//decode json request body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		message := fmt.Sprintf("JSON request decode failed, error: %s", err.Error())
		log.Error(message)
		jsonMessage, _ := json.Marshal(message)
		w.Write([]byte(jsonMessage))
		return
	}
	defer r.Body.Close()

	//validate the credentials received from request body
	correctPassword, ok := users[credentials.Username]
	if !ok || correctPassword != credentials.Password {
		w.WriteHeader(http.StatusUnauthorized)
		log.Error("password validation failed")
		w.Write([]byte("You are not Authorized"))
		return
	}
	expirationTime := time.Now().Add(time.Minute * 5)

	//creating payload to store username and expiry inside the jwt
	claims := &Claims{
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	//creating token with claims object and signing method as HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("error creating token")
		w.Write([]byte("Error logging in"))
		return
	}

	//setting token to a cookie
	http.SetCookie(w, /*cookie:=*/ &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	// err = cookie.Valid()
	// if err != nil {
	// 	log.Error("invalid cookie: %v", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// http.SetCookie(w, cookie)

	message := "successfully created Token"
	jsonMessage, _ := json.Marshal(message)
	log.Info(message)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonMessage))

}

func validatetoken(r *http.Request) bool {

	//get token from cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Error("No cookie found")
			return false
		}
		log.Error("error fetching token from cookie")
		return false
	}
	//storing token value to a variable
	tokenStr := cookie.Value
	claims := &Claims{}
	//parsing the token
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			log.Error("Invalid Signature")
			return false
		}
		log.Error("error parsing token")
		return false
	}
	//validating the token
	if !tkn.Valid {
		log.Error("token not valid")
		return false
	}
	return true
}

func InsertEmpDetails(w http.ResponseWriter, r *http.Request) {

	//validate token cookie
	ok := validatetoken(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		message := "Unauthorized access"
		log.Error(message)
		w.Write([]byte(message))
	}
	w.Header().Set("Content-Type", "application/json")
	var req InsertEmployeeRequest

	//decode request body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		message := fmt.Sprintf("JSON request decode failed, error: %s", err.Error())
		log.Error(message)
		jsonMessage, _ := json.Marshal(message)
		w.Write([]byte(jsonMessage))
		return
	}
	defer r.Body.Close()

	//execute sql query
	query := `INSERT INTO Employee (EmployeeID,LastName,FirstName,City)	VALUES (?,?,?,?)`
	_, err = DB.Exec(query, req.EmployeeID, req.LastName, req.FirstName, req.City)
	if err != nil {
		message := "failed to insert into db"
		log.Error(message, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(message))
		return
	}
	message := "successfully inserted employee details"
	jsonMessage, _ := json.Marshal(message)
	log.Info(message)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonMessage))

}

func GetEmpDetails(w http.ResponseWriter, r *http.Request) {

	//validate cookie token
	ok := validatetoken(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		message := "Unauthorized access"
		log.Error(message)
		w.Write([]byte(message))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var req GetEmpDetailsRequest

	//decode request body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		message := fmt.Sprintf("JSON request decode failed, error: %s", err.Error())
		jsonMessage, _ := json.Marshal(message)
		log.Error(message)
		w.Write([]byte(jsonMessage))
		return
	}
	defer r.Body.Close()

	query := `SELECT * FROM Employee WHERE EmployeeID = ?`
	row := DB.QueryRow(query, req.EmployeeID)

	var result []GetEmployeeResponseStruct
	value := GetEmployeeResponseStruct{}

	err = row.Scan(
		&value.EmployeeID,
		&value.LastName,
		&value.FirstName,
		&value.City,
	)
	if err != nil {
		message := "error scanning rows"
		log.Error(message, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(message))
		return
	}

	result = append(result, value)

	responsebody, _ := json.Marshal(result)
	log.Info("Getemployee request successfully executed")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responsebody))

}

func GetAllEmployeeDetails(w http.ResponseWriter, r *http.Request) {

	//validate cookie token
	ok := validatetoken(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		message := "Unauthorized access"
		log.Error(message)
		w.Write([]byte(message))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	query := `SELECT * FROM Employee `
	rows, err := DB.Query(query)
	if err != nil {
		message := "failed to fetch from db"
		log.Error(message, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(message))
		return
	}
	var result []GetEmployeeResponseStruct
	value := GetEmployeeResponseStruct{}

	for rows.Next() {
		err = rows.Scan(
			&value.EmployeeID,
			&value.LastName,
			&value.FirstName,
			&value.City,
		)
		if err != nil {
			message := "error scanning rows"
			log.Error(message, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(message))
			return
		}

		result = append(result, value)
	}
	responsebody, _ := json.Marshal(result)
	log.Info("Get all employees request successfully executed")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responsebody))

}

func DeleteEmployeeRecord(w http.ResponseWriter, r *http.Request) {
	//validate token cookie
	ok := validatetoken(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		message := "Unauthorized access"
		log.Error(message)
		w.Write([]byte(message))
	}
	w.Header().Set("Content-Type", "application/json")
	var req DeleteEmpDetailsRequest

	//decode request body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		message := fmt.Sprintf("JSON request decode failed, error: %s", err.Error())
		log.Error(message)
		jsonMessage, _ := json.Marshal(message)
		w.Write([]byte(jsonMessage))
		return
	}
	defer r.Body.Close()

	//execute sql query
	query := `DELETE FROM Employee WHERE EmployeeID = ?`
	_, err = DB.Exec(query, req.EmployeeID)
	if err != nil {
		message := "failed to delete from db"
		log.Error(message, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(message))
		return
	}
	message := fmt.Sprintf("successfully deleted employee details for employeeID - %d", req.EmployeeID)
	jsonMessage, _ := json.Marshal(message)
	log.Info(message)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonMessage))
}
