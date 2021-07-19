package controllers

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/kamva/mgm/v3"
	"github.com/macsko/todolist/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

// Register user in database
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var userLogin models.UserInput
	if err := getJSON(w, r, &userLogin); err != nil {
		return
	}

	// Check uniqueness of username
	var isUsernameTaken models.User
	err := mgm.Coll(userModel).First(bson.M{"username": userLogin.Username}, &isUsernameTaken)
	if err == nil {
		http.Error(w, "username is taken", http.StatusUnprocessableEntity)
		return
	}else if err != mongo.ErrNoDocuments {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Password encryption
	var user models.User
	user.Username = userLogin.Username
	user.Password, err = bcrypt.GenerateFromPassword([]byte(userLogin.Password), 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Saving user to database
	err = mgm.Coll(userModel).Create(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccess(w)
}

// Login user to service and storing session in cookie
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var userLogin models.UserInput
	if err := getJSON(w, r, &userLogin); err != nil {
		return
	}

	// Checking correctness of username and password
	var userDatabase models.User
	err := mgm.Coll(userModel).First(bson.M{"username": userLogin.Username}, &userDatabase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "incorrect username", http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword(userDatabase.Password, []byte(userLogin.Password))
	if err != nil {
		http.Error(w, "incorrect password", http.StatusUnprocessableEntity)
		return
	}

	// Creating jwt token and cookie
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer: userLogin.Username,
		ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
	})

	token, err := claims.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:  "jwt",
		Value: token,
		Expires: time.Now().Local().Add(time.Hour * time.Duration(24)),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	sendSuccess(w)
}

// Removing user's cookie
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	// Outdating the cookie
	cookie := &http.Cookie{
		Name:  "jwt",
		Value: "",
		MaxAge: -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	sendSuccess(w)
}

// Authentication of user helper
func AuthenticateRequest(r *http.Request) (string, error) {
	// Getting the cookie
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return "", err
	}

	// Parsing token and validation
	token, err := jwt.ParseWithClaims(
		cookie.Value,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("SECRET_KEY")), nil
		},
	)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", errors.New("invalid token")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return "", errors.New("outdated token")
	}

	return claims.Issuer, nil
}

// Authentication of user
func AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	_, err := AuthenticateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	sendSuccess(w)
}

// Authentication middleware
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, err := AuthenticateRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		mux.Vars(r)["username"] = username
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
