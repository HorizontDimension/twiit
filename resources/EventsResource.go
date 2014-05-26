package resources

import (
	"labix.org/v2/mgo"

	"net/http"
	"time"

	"github.com/emicklei/go-restful"

	"twiit/models"
)

type Event struct {
	Session *mgo.Session
}

type Calendar struct {
	Success int              `json:"success"`
	Result  []CalendarResult `json:"result"`
}

type CalendarResult struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	Class       string `json:"class"`
	Start       int64  `json:"start"`
	End         int64  `json:"end"`
	Description string `json:"description"`
	Photo       string `json:"photo"`
}

func (e Event) Register(container *restful.Container) {

	ws := new(restful.WebService)
	ws.
		Path("/events").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{event-id}").To(e.GetEvent).
		// docs
		Doc("get a event").
		Operation("GetEvent").
		Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")).
		Writes(models.Events{})) // on the response

	ws.Route(ws.GET("/calendar").To(e.Calendar).
		// docs
		Doc("get events for calendar").
		Operation("GetEvent").
		Writes(Calendar{})) // on the response

	ws.Route(ws.GET("/latests").To(e.Latests).
		// docs
		Doc("get latest events").
		Operation("Latests").
		Writes([]*models.Events{})) // on the response

	ws.Route(ws.POST("").To(e.CreateEvent).
		// docs
		Doc("create a event").
		Operation("PostUser").
		Reads(models.Events{})) // from the request

	ws.Route(ws.PUT("/{event-id}").To(e.UpdateEvent).
		// docs
		Doc("update a event").
		Operation("PutUser").
		Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")).
		Reads(models.Events{})) // from the request

	ws.Route(ws.DELETE("/{event-id}").To(e.DeleteEvent).
		// docs
		Doc("delete a event").
		Operation("DeleteEvent").
		Param(ws.PathParameter("event-id", "identifier of the event").DataType("string")))

	ws.Route(ws.GET("/search/{search-term}").To(e.SearchEvent).
		// docs
		Doc("Search a event").
		Operation("SearchEvent").
		Param(ws.PathParameter("search-term", "term to search").DataType("string")).
		Writes([]models.Events{})) // on the response
	container.Add(ws)
}

func (e *Event) GetAllEvents(request *restful.Request, response *restful.Response) {
	events := models.GetAllEvents(e.Session)
	response.WriteEntity(events)
}

func (e *Event) Latests(request *restful.Request, response *restful.Response) {
	levents := models.GetLatestEvents(e.Session, 3)
	response.WriteEntity(levents)
}

func (e *Event) Calendar(request *restful.Request, response *restful.Response) {

	var events []*models.Events

	events = models.GetAllEvents(e.Session)

	result := Calendar{}
	result.Success = 0

	if len(events) > 0 {
		result.Success = 1
	}

	for i, event := range events {
		cr := CalendarResult{}
		cr.Start = event.Date.UTC().Unix() * 1000
		cr.Title = event.Name
		cr.Url = "/events/" + event.Id.Hex()
		cr.End = event.Date.Add(8*time.Hour).UTC().Unix() * 1000
		cr.Id = i
		cr.Photo = event.Thumb.Hex()
		cr.Description = event.Description
		cr.Class = event.Priority
		result.Result = append(result.Result, cr)
	}

	response.WriteEntity(result)

}

func (e *Event) GetEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	event := models.GetEventById(e.Session, id)
	response.WriteEntity(event)
}

//insert a user
func (e *Event) CreateEvent(request *restful.Request, response *restful.Response) {
	var event models.Events
	err := request.ReadEntity(&event)
	if err != nil {
		if err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}
	}

	//todo validate entry

	err = event.Save(e.Session)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

}

//update a user
func (e *Event) UpdateEvent(request *restful.Request, response *restful.Response) {

}

//search a user
func (e *Event) SearchEvent(request *restful.Request, response *restful.Response) {
	searchterm := request.PathParameter("search-term")
	events := models.FindEvents(e.Session, searchterm, 20)
	response.WriteEntity(events)
}

func (e *Event) DeleteEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	err := models.EventCol(e.Session).RemoveId(id)

	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

}
