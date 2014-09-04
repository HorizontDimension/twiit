package twiit

import (
	"net/http"
	"time"

	"github.com/HorizontDimension/jwt-go"
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
func (t *Token) Generate() (tk string, err error) {
	tk, err = t.token.SignedString(t.key)
	if err == nil {
		CacheAuth.Set(t.Get("id").(string), tk, t.duration)

	}

	return
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
		Log.Warn("fail to parse jwt token", "error", err)
		return nil, err
	}

	return t, nil
}

func ParseTokenFromReq(req *http.Request) (*Token, error) {

	t, err := jwt.ParseFromRequest(req, func(token *jwt.Token) ([]byte, error) {
		//_, ok := Cache.Get(token.Claims["id"].(string))
		//if !ok {
		//		return []byte(""), nil
		//	}

		//todo implement extra validation
		return secretKey, nil
	})
	if err != nil {
		Log.Warn("fail to parse jwt token from request", "error", err)
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
		Log.Error("fail to generate jwt token", "error", err)
		return err
	}
	respstring := `{ "token": "` + stoken + `" }`
	_, err = rw.Write([]byte(respstring))
	if err != nil {
		Log.Error("failed to write generated token in RespomseWriter", "error", err)
	}
	return nil
}

func (t *Token) IsValid() bool {
	return t.token.Valid
}
