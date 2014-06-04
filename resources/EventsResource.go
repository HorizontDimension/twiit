package resources

import (
	"net/http"
	"time"

	"strconv"

	"github.com/HorizontDimension/twiit"
	"github.com/HorizontDimension/twiit/models"
	"github.com/emicklei/go-restful"
	"labix.org/v2/mgo"
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

	ws.Route(ws.GET("/latests/{number-events}").To(e.Latests).
		// docs
		Doc("get latest  events").
		Param(ws.PathParameter("number-events", "Number of events to get").DataType("string")).
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
	err := response.WriteEntity(events)
	if err != nil {
		twiit.Log.Error("Error writing response on GetAllEvents ", "error", err)

	}
}

func (e *Event) Latests(request *restful.Request, response *restful.Response) {
	numbers := request.PathParameter("number-events")
	number, err := strconv.Atoi(numbers)
	if err != nil {
		twiit.Log.Error("Error converting number-events to int", "error", err)
	}

	levents := models.GetLatestEvents(e.Session, number)
	err = response.WriteEntity(levents)
	if err != nil {
		twiit.Log.Error("Error writing response on Latests ", "error", err)

	}
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

	err := response.WriteEntity(result)
	if err != nil {
		twiit.Log.Error("Error writing response on Calendar ", "error", err)

	}

}

func (e *Event) GetEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	event := models.GetEventById(e.Session, id)
	err := response.WriteEntity(event)
	if err != nil {
		twiit.Log.Error("Error writing response on GetEvent ", "error", err)

	}

}

//insert a user
func (e *Event) CreateEvent(request *restful.Request, response *restful.Response) {
	var event models.Events
	err := request.ReadEntity(&event)
	if err != nil {
		twiit.Log.Error("Error reading Entity from request on  CreateEvent", "error", err)

		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusInternalServerError, err.Error())
		if err != nil {
			twiit.Log.Error("Error writing response on CreateEvent ", "error", err)

		}
		return

	}

	//todo validate entry

	err = event.Save(e.Session)
	if err != nil {
		twiit.Log.Error("Error Saving Event on  CreateEvent", "event", event, "error", err)
		response.AddHeader("Content-Type", "text/plain")
		err = response.WriteErrorString(http.StatusInternalServerError, err.Error())
		if err != nil {
			twiit.Log.Error("Error writing response on CreateEvent ", "error", err)

		}
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
	err := response.WriteEntity(events)
	if err != nil {
		twiit.Log.Error("Error writing response on SearchEvent ", "error", err)

	}
}

func (e *Event) DeleteEvent(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("event-id")
	err := models.EventCol(e.Session).RemoveId(id)
	twiit.Log.Error("Error Deleting Event on  CreateEvent", "eventid", id, "error", err)

	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusInternalServerError, err.Error())
		if err != nil {
			twiit.Log.Error("Error writing response on DeleteEvent ", "error", err)

		}

		return
	}

}
