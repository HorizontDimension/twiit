package resources

import (
	"fmt"

	"net/http"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/HorizontDimension/twiit/models"
	"github.com/emicklei/go-restful"
)

type Guest struct {
	Session *mgo.Session
}

func buildSearchFromUsers(users []*models.User) (results []models.SearchResult) {

	for _, user := range users {
		result := models.SearchResult{}

		result.Tokens = user.Tokens
		switch user.Role {
		case models.UserClient:
			result.Kind = "user"
			result.IsGuest = true
			result.Id = user.Id.Hex()
			result.EditUrl = "/#/users/edit/" + user.Id.Hex()
		case models.UserPromotor:
			result.Kind = "glass"
			result.EditUrl = "/#/promotors/edit/" + user.Id.Hex()
			result.Id = user.Id.Hex()
		}
		result.Image = "/@admin/public/img/user.jpg"
		if bson.IsObjectIdHex(user.Photo.Hex()) {
			result.Image = user.Thumb.Hex()
		}
		result.Url = "/#/users/read/" + user.Id.Hex()
		result.Value = user.Firstname + " " + user.Lastname
		results = append(results, result)
	}

	return results

}

func (u Guest) Register(container *restful.Container) {

	ws := new(restful.WebService)
	ws.
		Path("/guest").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{user-id}").Filter(RequirePromotor).To(u.GetGuest).
		// docs
		Doc("get a guest").
		Operation("GetUser").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Writes(models.User{})) // on the response

	ws.Route(ws.POST("").Filter(RequirePromotor).To(u.CreateGuest).
		// docs
		Doc("create a guest").
		Operation("CreateGuest").
		Reads(models.User{})) // from the request

	ws.Route(ws.PUT("/{user-id}").Filter(RequirePromotor).To(u.UpdateGuest).
		// docs
		Doc("update a guest").
		Operation("UpdateGuest").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Reads(models.User{})) // from the request

	ws.Route(ws.DELETE("/{user-id}").Filter(RequirePromotor).To(u.DeleteGuest).
		// docs
		Doc("delete a guest").
		Operation("DeleteGuest").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

	ws.Route(ws.GET("/search/{search-term}").Filter(RequirePromotor).To(u.SearchGuest).
		// docs
		Doc("Search a user").
		Operation("SearchGuest").
		Param(ws.PathParameter("search-term", "term to search").DataType("string")).
		Writes(models.SearchResult{})) // on the response

	container.Add(ws)
}

func (u *Guest) GetAllGuests(request *restful.Request, response *restful.Response) {

}

func (u *Guest) GetGuest(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	user := models.GetUserById(u.Session, id)
	response.WriteEntity(user)
}

//insert a user
func (u *Guest) CreateGuest(request *restful.Request, response *restful.Response) {
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
	user.Role = models.UserClient
	err = user.Save(u.Session)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(user)
}

//update a user
func (u *Guest) UpdateGuest(request *restful.Request, response *restful.Response) {
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

//search a user
func (u *Guest) SearchGuest(request *restful.Request, response *restful.Response) {
	searchterm := request.PathParameter("search-term")
	users := buildSearchFromUsers(models.FindUser(u.Session, searchterm, models.UserClient, 20))
	if users != nil {
		response.WriteEntity(users)
		return
	}
	//angular expects an array
	response.Write([]byte("[]"))

}

func (u *Guest) DeleteGuest(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	users := models.GetUserById(u.Session, id)
	users.Delete(u.Session)

}
