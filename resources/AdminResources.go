package resources

import (
	"fmt"

	//"io/ioutil"
	"net/http"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/emicklei/go-restful"
	"twiit/models"
)

type Admin struct {
	Session *mgo.Session
}

func (p *Admin) AssociateGuest(guestId bson.ObjectId) {

}

func (u Admin) Register(container *restful.Container) {

	ws := new(restful.WebService)
	ws.
		Path("/admin").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{user-id}").Filter(RequireAdmin).To(u.GetAdmin).
		// docs
		Doc("get a Admin").
		Operation("GetAdmin").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Writes(models.User{})) // on the response

	ws.Route(ws.POST("").To(u.CreateAdmin).
		// docs
		Doc("create a Admin").
		Operation("CreateAdmin").
		Reads(models.User{})) // from the request

	ws.Route(ws.PUT("/{user-id}").To(u.UpdateAdmin).
		// docs
		Doc("update a Admin").
		Operation("UpdateAdmin").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Reads(models.User{})) // from the request

	ws.Route(ws.DELETE("/{user-id}").To(u.DeleteAdmin).
		// docs
		Doc("delete a Admin").
		Operation("DeleteAdmin").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

	ws.Route(ws.GET("/search/{search-term}").Filter(RequireAdmin).To(u.SearchAdmin).
		// docs
		Doc("Search a promotor").
		Operation("SearchAdmin").
		Param(ws.PathParameter("search-term", "term to search").DataType("string")).
		Writes(models.SearchResult{})) // on the response

	container.Add(ws)
}

//+admin
func (u *Admin) GetAllAdmins(request *restful.Request, response *restful.Response) {

}

//+admin
func (u *Admin) GetAdmin(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	user := models.GetUserById(u.Session, id)
	response.WriteEntity(user)
}

//+admin
//insert a user
func (u *Admin) CreateAdmin(request *restful.Request, response *restful.Response) {

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
	user.Role = models.UserAdmin
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
func (u *Admin) UpdateAdmin(request *restful.Request, response *restful.Response) {
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
func (u *Admin) SearchAdmin(request *restful.Request, response *restful.Response) {
	searchterm := request.PathParameter("search-term")
	users := buildSearchFromUsers(models.FindUser(u.Session, searchterm, models.UserAdmin, 20))
	if users != nil {
		response.WriteEntity(users)
		return
	}
	//angular expects an array
	response.Write([]byte("[]"))

}

//+admin
func (u *Admin) DeleteAdmin(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	users := models.GetUserById(u.Session, id)
	users.Delete(u.Session)

}
