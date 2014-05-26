package resources

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/emicklei/go-restful"
	"labix.org/v2/mgo"
	"net/http"
	"twiit"
	"twiit/models"
	//"twiit/utils"

	"log"
)

type Auth struct {
	Session *mgo.Session
}

type authform struct {
	Email    string
	Password string
}

func (a *Auth) Register(container *restful.Container) {

	ws := new(restful.WebService)
	ws.
		Path("/auth").
		Consumes("*/*").
		Produces("*/*") // you can specify this per route as well

	ws.Route(ws.POST("login").Consumes(restful.MIME_JSON).To(a.Login).
		// docs
		Doc("Authenticate a user").
		Operation("Login").
		Reads(authform{})) // on the response

	container.Add(ws)
}

func (a *Auth) Login(request *restful.Request, response *restful.Response) {

	var userauth authform
	err := request.ReadEntity(&userauth)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	user := models.GetUserByEmail(a.Session, userauth.Email)

	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(userauth.Password))

	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteError(http.StatusInternalServerError, err)
		return

	}

	token := twiit.NewToken()

	token.Set("id", user.Id.Hex())
	token.Set("role", user.Role)
	token.WriteHeader(response.ResponseWriter)
	s, err := token.Generate()
	log.Println("created", s, err)
}

func RequirePromotor(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		resp.WriteErrorString(http.StatusUnauthorized, "not authorized: unable to parse token")
		return
	}

	if !token.IsValid() {
		resp.WriteErrorString(http.StatusUnauthorized, "not authorized")
		return

	}

	role := uint8(token.Get("role").(float64))
	if role != models.UserPromotor && role != models.UserAdmin {
		log.Println("----------->", role, models.UserPromotor)
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	chain.ProcessFilter(req, resp)
}

func RequireAdmin(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	if !token.IsValid() {
		resp.WriteErrorString(http.StatusUnauthorized, "not authorized")
		return

	}
	role := token.Get("role").(uint8)
	if role != models.UserAdmin {
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	chain.ProcessFilter(req, resp)
}

func RequireDoorman(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}
	if !token.IsValid() {
		resp.WriteErrorString(http.StatusUnauthorized, "not authorized")
		return

	}
	role := token.Get("role").(uint8)
	if role != models.UserGateKeeper && role != models.UserAdmin {
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	chain.ProcessFilter(req, resp)
}

func RequireLoggedUser(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		resp.WriteErrorString(http.StatusUnauthorized, "not authorized: unable to parse token")
		return
	}

	if !token.IsValid() {
		resp.WriteErrorString(http.StatusUnauthorized, "not authorized not valid token")
		return

	}
	role := (token.Get("role")).(uint8)

	log.Println(role == models.UserClient)

	if role != models.UserGateKeeper && role != models.UserAdmin && role != models.UserPromotor && role != models.UserClient {

		resp.WriteErrorString(http.StatusUnauthorized, "not authorized not valid user")
		return
	}

	chain.ProcessFilter(req, resp)
}
