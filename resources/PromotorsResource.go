package resources

import (
	"fmt"

	"net/http"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/HorizontDimension/twiit"
	"github.com/HorizontDimension/twiit/models"
	"github.com/emicklei/go-restful"
)

type Promotor struct {
	Session *mgo.Session
}

func (u Promotor) Register(container *restful.Container) {

	ws := new(restful.WebService)
	ws.
		Path("/promotor").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{user-id}").Filter(RequirePromotor).To(u.GetPromotor).
		// docs
		Doc("get a Promotor").
		Operation("GetPromotor").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Writes(models.User{})) // on the response

	ws.Route(ws.POST("").To(u.CreatePromotor).
		// docs
		Doc("create a Promotor").
		Operation("CreatePromotor").
		Reads(models.User{})) // from the request

	ws.Route(ws.POST("/uninviteguest/{guest-id}").To(u.UninviteGuest).
		// docs
		Doc("uninvite a guest").
		Operation("uninviteguest").
		Param(ws.PathParameter("guest-id", "identifier of the guest").DataType("string")).
		Reads("event-id")) // from the request

	ws.Route(ws.POST("/inviteguest/{guest-id}").To(u.InviteGuest).
		// docs
		Doc("Invite a guest").
		Operation("InviteGuest").
		Param(ws.PathParameter("guest-id", "identifier of the guest").DataType("string")).
		Reads("event-id")) // from the request

	ws.Route(ws.PUT("/{user-id}").To(u.UpdatePromotor).
		// docs
		Doc("update a Promotor").
		Operation("UpdatePromotor").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Reads(models.User{})) // from the request

	ws.Route(ws.DELETE("/{user-id}").To(u.DeletePromotor).
		// docs
		Doc("delete a Promotor").
		Operation("DeletePromotor").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

	ws.Route(ws.GET("/search/{search-term}").Filter(RequirePromotor).To(u.SearchPromotor).
		// docs
		Doc("Search a promotor").
		Operation("SearchPromotor").
		Param(ws.PathParameter("search-term", "term to search").DataType("string")).
		Writes(models.SearchResult{})) // on the response

	container.Add(ws)
}

//todo validate user input
func (p *Promotor) InviteGuest(request *restful.Request, response *restful.Response) {
	var guestid, eventid string

	guestid = request.PathParameter("guest-id")

	err := request.ReadEntity(&eventid)
	if err != nil {
		twiit.Log.Error("Error Reading event-id from request ", "error", err)
	}

	//validate event

	//get Event
	event := models.GetEventById(p.Session, eventid)
	if event == nil {
		twiit.Log.Warn("Event not found on InviteGuest", "error", err, "eventid", eventid)
		return
	}
	guest := models.GetUserById(p.Session, guestid)
	if event == nil {
		twiit.Log.Warn("User not found on InviteGuest", "error", err, "eventid", eventid)
		return
	}

	tk, err := twiit.ParseTokenFromReq(request.Request)
	if err != nil {
		twiit.Log.Info("error parsing token on inviteguest", "error", err)

	}

	//get guestlist for logged user(must be promotor)
	gl := event.GuestlistByOwner(bson.ObjectIdHex(tk.Get("id").(string)))
	gl.AddGuest(guest.Id)

	err = event.Save(p.Session)
	if err != nil {
		twiit.Log.Error("Error saving invite guest to DB", "error", err, "eventid", eventid, "guest", tk.Get("id").(string))
	}

}

//todo validate user input
func (p *Promotor) UninviteGuest(request *restful.Request, response *restful.Response) {
	var guestid, eventid string

	guestid = request.PathParameter("guest-id")

	err := request.ReadEntity(&eventid)
	if err != nil {
		twiit.Log.Error("Error Reading event-id from request ", "error", err)
	}

	//validate event

	//get Event
	event := models.GetEventById(p.Session, eventid)
	if event == nil {
		twiit.Log.Warn("Event not found on InviteGuest", "error", err, "eventid", eventid)
		return
	}
	guest := models.GetUserById(p.Session, guestid)
	if event == nil {
		twiit.Log.Warn("User not found on InviteGuest", "error", err, "eventid", eventid)
		return
	}

	tk, err := twiit.ParseTokenFromReq(request.Request)
	if err != nil {
		twiit.Log.Info("error parsing token on inviteguest", "error", err)

	}

	gl := event.GuestlistByOwner(bson.ObjectIdHex(tk.Get("id").(string)))
	gl.RemoveGuest(guest.Id)

	err = event.Save(p.Session)
	if err != nil {
		twiit.Log.Error("Error saving uninvite guest to DB", "error", err, "eventid", eventid, "guest", tk.Get("id").(string))
	}
}

//+admin
func (u *Promotor) GetAllPromotors(request *restful.Request, response *restful.Response) {

}

//+admin
func (u *Promotor) GetPromotor(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	user := models.GetUserById(u.Session, id)
	err := response.WriteEntity(user)
	if err != nil {
		twiit.Log.Error("Error writing response on GetPromotor ", "error", err)
	}
}

//+admin
//insert a user
func (u *Promotor) CreatePromotor(request *restful.Request, response *restful.Response) {
	user := models.User{}
	err := request.ReadEntity(&user)
	if err != nil {
		twiit.Log.Error("Error Reading Entity request on CreatePromotor ", "error", err)

		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusInternalServerError, err.Error())
		if err != nil {
			twiit.Log.Error("Error writing response on CreatePromotor ", "error", err)
		}
		return
	}

	if exists := models.GetUserByEmail(u.Session, user.Email); exists.Email == user.Email {
		msg := fmt.Sprint("Account with ", user.Email, " already exists.")
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusBadRequest, msg)
		if err != nil {
			twiit.Log.Error("Error writing response on CreatePromotor ", "error", err)
		}
		return
	}

	user.Id = bson.NewObjectId()
	user.Role = models.UserPromotor
	err = user.Save(u.Session)
	if err != nil {
		twiit.Log.Error("Error saving Promotor on CreatePromotor ", "error", err, "promotor", user)
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusInternalServerError, err.Error())
		if err != nil {
			twiit.Log.Error("Error writing response on CreatePromotor ", "error", err)
		}
		return
	}

	response.WriteHeader(http.StatusCreated)
	err = response.WriteEntity(user)
	if err != nil {
		twiit.Log.Error("Error writing response on CreatePromotor ", "error", err)
	}
}

//+admin || self promotor
//update a user
func (u *Promotor) UpdatePromotor(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")

	ou := models.GetUserById(u.Session, id)
	user := new(models.User)
	user.Id = ou.Id
	user.Role = ou.Role
	user.Created = ou.Created

	if exists := models.GetUserByEmail(u.Session, user.Email); exists.Email == user.Email && exists.Id != user.Id {
		msg := fmt.Sprint("Account with ", user.Email, " already exists.")
		twiit.Log.Info(msg, "email", user.Email)
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusBadRequest, msg)
		if err != nil {
			twiit.Log.Error("Error writing response on UpdatePromotor ", "error", err)
		}

		return
	}

	//todo chek error
	err := user.Save(u.Session)
	if err != nil {
		twiit.Log.Error("Error saving promotor on UpdatePromotor ", "error", err, "promotor", user)
	}
	err = response.WriteEntity(user)
	if err != nil {
		twiit.Log.Error("Error writing response on UpdatePromotor ", "error", err)
	}
}

//all?
//search a user
func (u *Promotor) SearchPromotor(request *restful.Request, response *restful.Response) {
	searchterm := request.PathParameter("search-term")
	users := buildSearchFromUsers(models.FindUser(u.Session, searchterm, models.UserClient, 20))
	if users != nil {
		err := response.WriteEntity(users)
		if err != nil {
			twiit.Log.Error("Error writing response on SearchPromotor ", "error", err)
		}

		return
	}
	//angular expects an array
	_, err := response.Write([]byte("[]"))
	if err != nil {
		twiit.Log.Error("Error writing response on SearchPromotor ", "error", err)
	}

}

//+admin
func (u *Promotor) DeletePromotor(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	users := models.GetUserById(u.Session, id)
	err := users.Delete(u.Session)
	if err != nil {
		twiit.Log.Error("Error deleting promotor on DeletePromotor ", "error", err, "promotorid", id)
	}

}

//+admin
func CreatePromotorFromUser() {

}
