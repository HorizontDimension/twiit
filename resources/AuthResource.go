package resources

import (
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"

	"github.com/HorizontDimension/twiit"
	"github.com/HorizontDimension/twiit/models"
	"github.com/emicklei/go-restful"

	"labix.org/v2/mgo"
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
		twiit.Log.Error("Error reading auth entity on Auth.Login ", "error", err)

		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			twiit.Log.Error("Error writing response on Auth.Login ", "error", err)

		}
		return
	}

	user := models.GetUserByEmail(a.Session, userauth.Email)

	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(userauth.Password))

	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteError(http.StatusInternalServerError, err)
		if err != nil {
			twiit.Log.Error("Error writing response on Auth.Login ", "error", err)

		}
		return

	}

	token := twiit.NewToken()

	token.Set("id", user.Id.Hex())
	token.Set("role", user.Role)
	err = token.WriteHeader(response.ResponseWriter)
	if err != nil {
		twiit.Log.Error("Error writing response on Auth.Login ", "error", err)
	}

}

func RequirePromotor(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		twiit.Log.Warn("Error parsing token ", "error", err)
		err = resp.WriteErrorString(http.StatusUnauthorized, "not authorized")
		if err != nil {
			twiit.Log.Error("Error writing response on RequirePromotor ", "error", err)

		}
		return
	}

	if !token.IsValid() {
		err = resp.WriteErrorString(http.StatusUnauthorized, "not authorized")
		if err != nil {
			twiit.Log.Error("Error writing response on RequirePromotor ", "error", err)

		}
		return

	}

	//get rid of that compiller bug??
	role := uint8(token.Get("role").(float64))

	if role != models.UserPromotor && role != models.UserAdmin {
		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequirePromotor ", "error", err)

		}
		return
	}

	chain.ProcessFilter(req, resp)
}

func RequireAdmin(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		err = resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireAdmin ", "error", err)

		}
		return
	}

	if !token.IsValid() {
		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireAdmin ", "error", err)

		}
		return

	}
	role := token.Get("role").(uint8)
	if role != models.UserAdmin {
		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireAdmin ", "error", err)

		}
		return
	}

	chain.ProcessFilter(req, resp)
}

func RequireDoorman(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		twiit.Log.Info("Unable to parse token from request on RequireDoorman ", "error", err)
		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireDoorman ", "error", err)

		}
		return
	}
	if !token.IsValid() {
		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireDoorman ", "error", err)

		}
		return

	}
	role := token.Get("role").(uint8)
	if role != models.UserGateKeeper && role != models.UserAdmin {
		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")
		if err != nil {
			twiit.Log.Error("Error writing response on RequireDoorman ", "error", err)

		}
		return
	}

	chain.ProcessFilter(req, resp)
}

func RequireLoggedUser(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token, err := twiit.ParseTokenFromReq(req.Request)
	if err != nil {
		twiit.Log.Info("Unable to parse token from request on RequireLoggedUser ", "error", err)

		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireLoggedUser ", "error", err)

		}
		return
	}

	if !token.IsValid() {
		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireLoggedUser ", "error", err)

		}
		return

	}
	role := (token.Get("role")).(uint8)

	if role != models.UserGateKeeper && role != models.UserAdmin && role != models.UserPromotor && role != models.UserClient {

		err := resp.WriteErrorString(http.StatusUnauthorized, "not authorized")

		if err != nil {
			twiit.Log.Error("Error writing response on RequireLoggedUser ", "error", err)

		}
		return
	}

	chain.ProcessFilter(req, resp)
}
