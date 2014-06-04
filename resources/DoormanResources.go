package resources

import (
	"net/http"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/HorizontDimension/twiit"
	"github.com/HorizontDimension/twiit/models"
	"github.com/emicklei/go-restful"
)

type Doorman struct {
	Session *mgo.Session
}

type CheckIn struct {
	Guest    bson.ObjectId
	Promotor bson.ObjectId
	CartId   int
}

func (d *Doorman) Register(container *restful.Container) {

	ws := new(restful.WebService)
	ws.
		Path("/doorman").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.POST("/checkin/{event-id}").To(d.Checkin).
		Doc("check in a guest").
		Operation("CheckIn").
		Param(ws.PathParameter("event-id", "id of the event").DataType("string")).
		Reads(CheckIn{})) // on the request

	container.Add(ws)

}

func (d *Doorman) Checkin(request *restful.Request, response *restful.Response) {

	eventid := request.PathParameter("event-id")

	if eventid == "" {
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusBadRequest, "empty event id")
		if err != nil {
			twiit.Log.Error("Error writing response on Checkin ", "error", err)

		}
		return
	}
	var entry CheckIn

	err := request.ReadEntity(&entry)
	if err != nil {
		twiit.Log.Error("Error Reading cardid from request ", "error", err)
	}

	event := models.GetEventById(d.Session, eventid)

	//validate entrydata

	event.CheckInGuest(
		entry.Promotor,
		entry.Guest,
		entry.CartId)
	err = event.Save(d.Session)
	if err != nil {
		twiit.Log.Error("Error Saving Event on  Checkin", "event", event, "error", err)
		response.AddHeader("Content-Type", "text/plain")
		err = response.WriteErrorString(http.StatusInternalServerError, err.Error())
		if err != nil {
			twiit.Log.Error("Error writing response on Checkin ", "error", err)

		}
		return
	}
}
