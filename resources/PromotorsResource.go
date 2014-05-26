package resources

import (
	"fmt"

	"net/http"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/emicklei/go-restful"
	"github.com/HorizontDimension/twiit/models"
)

type Promotor struct {
	Session *mgo.Session
}

func (p *Promotor) AssociateGuest(guestId bson.ObjectId) {

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

//+admin
func (u *Promotor) GetAllPromotors(request *restful.Request, response *restful.Response) {

}

//+admin
func (u *Promotor) GetPromotor(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	user := models.GetUserById(u.Session, id)
	response.WriteEntity(user)
}

//+admin
//insert a user
func (u *Promotor) CreatePromotor(request *restful.Request, response *restful.Response) {
	user := models.User{}
	err := request.ReadEntity(&user)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if exists := models.GetUserByEmail(u.Session, user.Email); exists.Email == user.Email {
		msg := fmt.Sprint("Account with ", user.Email, " already exists.")
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, msg)
		return
	}

	user.Id = bson.NewObjectId()
	user.Role = models.UserPromotor
	err = user.Save(u.Session)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(user)
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
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, msg)

		return
	}

	//todo chek error
	user.Save(u.Session)
	response.WriteEntity(user)
}

//all?
//search a user
func (u *Promotor) SearchPromotor(request *restful.Request, response *restful.Response) {
	searchterm := request.PathParameter("search-term")
	users := buildSearchFromUsers(models.FindUser(u.Session, searchterm, models.UserClient, 20))
	if users != nil {
		response.WriteEntity(users)
		return
	}
	//angular expects an array
	response.Write([]byte("[]"))

}

//+admin
func (u *Promotor) DeletePromotor(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	users := models.GetUserById(u.Session, id)
	users.Delete(u.Session)
}

//+admin
func CreatePromotorFromUser() {

}
