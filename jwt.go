package twiit

import (
	//"errors"
	"github.com/HorizontDimension/jwt-go"
	"log"
	"net/http"
	"time"
)

var (
	secretKey = []byte("google")
	duration  = 72 * time.Hour
)

type Token struct {
	key      []byte
	token    *jwt.Token
	duration time.Duration
}

func NewToken() (t *Token) {

	t = new(Token)
	t.token = jwt.New(jwt.GetSigningMethod("HS256"))
	t.duration = duration
	t.key = secretKey
	t.token.Claims["exp"] = time.Now().Add(t.duration).Unix()
	return t

}

func (t *Token) Get(key string) interface{} {
	return t.token.Claims[key]
}

func (t *Token) Set(key string, value interface{}) {
	if key != "exp" {
		t.token.Claims[key] = value
	}
}
func (t *Token) Generate() (string, error) {
	return t.token.SignedString(t.key)
}

func ParseToken(toke string) (*Token, error) {
	t := new(Token)
	t.key = secretKey
	t.token = new(jwt.Token)
	var err error
	t.token, err = jwt.Parse(toke, func(token *jwt.Token) ([]byte, error) {

		//todo implement extra validation
		return t.key, nil
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

func ParseTokenFromReq(req *http.Request) (*Token, error) {

	t, err := jwt.ParseFromRequest(req, func(token *jwt.Token) ([]byte, error) {

		//todo implement extra validation
		return secretKey, nil
	})
	if err != nil {
		log.Println("...", err)
		return nil, err
	}
	token := new(Token)
	token.key = secretKey
	token.token = new(jwt.Token)
	token.token = t
	return token, nil

}

func (t *Token) WriteHeader(rw http.ResponseWriter) error {
	stoken, err := t.Generate()
	if err != nil {
		return err
	}
	respstring := `{ "token": "` + stoken + `" }`
	rw.Write([]byte(respstring))
	return nil
}

func (t *Token) IsValid() bool {
	return t.token.Valid
}
