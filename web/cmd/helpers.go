package main

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"net/http"
	"time"
)

const (
	idCookieName    = "id"
	tokenCookieName = "token"
	expDay          = 60 * 24
)

//Вощвращает токен, считанный из куки
func getTokenCookies(r *http.Request) *token {
	cookieId, err := r.Cookie(idCookieName)
	if err != nil {
		return nil
	}
	cookieToken, err := r.Cookie(tokenCookieName)
	if err != nil {
		return nil
	}
	if cookieId.Value == "" || cookieToken.Value == "" {
		return nil
	}
	return &token{
		IdUser: cookieId.Value,
		Token:  cookieToken.Value,
	}
}

//Возвращает новый объект куки
func newCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:    name,
		Value:   value,
		Path:    "/",
		Expires: time.Now().Add(expDay * time.Hour),
	}
}

//Функция авторизации пользователя
//Ищет совпадения в базе пользователей
//Выдает новый токен доступа
//при успехе возвращается пустая строка
func auth(w http.ResponseWriter, email, password string) error {
	u := getUserByEmail(email)
	if u == nil {
		return errors.New("user not found")
	}
	err := u.comparePassword(password)
	if err != nil {
		return err
	}
	genToken, err := generateToken(u.Id.Hex())
	if err != nil {
		return err
	}
	tkn := token{
		IdUser: u.Id.Hex(),
		Token:  genToken,
	}
	err = tkn.saveToken(w)
	if err != nil {
		return err
	}
	return nil
}

//Генерирует новый токен на основе какого-то слова
func generateToken(word string) (string, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	n := 20
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	bcryptB, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return word + string(time.Now().Unix()) + string(bcryptB), nil
}

//Проверка токена доступа, возвращает токен с данными при успехе
func checkAuth(r *http.Request) (*token, *user) {
	tkn := getTokenCookies(r)
	if tkn == nil {
		return nil, nil
	}
	is := tkn.findInDB()
	if !is {
		return nil, nil
	}
	u := getUserById(bson.ObjectIdHex(tkn.IdUser))
	if u == nil {
		return nil, nil
	}
	return tkn, u
}
