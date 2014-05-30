package models

import (
	"strings"
	"time"
	//"unsafe"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/HorizontDimension/twiit"
	"github.com/HorizontDimension/twiit/utils"
)

func EventCol(s *mgo.Session) *mgo.Collection {
	return s.DB("twiit").C("Event")
}

type EventEntries struct {
	Entrytime  time.Time
	Event      bson.ObjectId
	Promotor   bson.ObjectId
	Guest      bson.ObjectId
	CardNumber bson.ObjectId
}

type Events struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	Name        string
	Description string
	Date        time.Time
	Tags        string
	Priority    string
	Image       bson.ObjectId
	Thumb       bson.ObjectId
	GuestList   []GuestList
	Tokens      []string
}

func (e *Events) HasGuestlist(s *mgo.Session, owner bson.ObjectId) (event *Events) {
	query := bson.M{"_id": e.Id, "guestlist": bson.M{"$elemMatch": bson.M{"owner": owner}}}
	err := EventCol(s).Find(query).One(event)
	if err != nil {
		twiit.Log.Error("failed to find guestlist", "error", err, "query", query)
	}
	return event
}

//Guestlist returns a existing guestlist for a given user or create one if doesnt exist
func (e *Events) GuestlistByOwner(owner bson.ObjectId) *GuestList {
	var index int
	var found = false

	if len(e.GuestList) < 1 {
		e.GuestList = []GuestList{}
		e.GuestList = append(e.GuestList, *(NewGuestlist(owner, "")))
		return &e.GuestList[0]
	}

	for i := range e.GuestList {
		if e.GuestList[i].Owner == owner {
			found = true
			index = i
			break
		}
	}

	if !found {
		e.GuestList = []GuestList{}
		e.GuestList = append(e.GuestList, *(NewGuestlist(owner, "")))
		for i := range e.GuestList {
			if e.GuestList[i].Owner == owner {
				found = true
				index = i
				break
			}
		}
	}

	if found {
		return &e.GuestList[index]
	}
	return nil
}

func (e *Events) AddToGuestlist(owner bson.ObjectId, guest bson.ObjectId) {

	//guestlist := e.GuestlistByOwner(owner)
	//	guestlist.AddGuest(guest)
	//	err := e.Save(e.)
	//if err != nil {
	//		twiit.Log.Error("error saving event in AddToGuestlist", "error", err, "owner", owner, "guest", guest)
	//}
}

func (e *Events) buildTokenList() {
	e.Tokens = []string{}
	e.Tokens = append(e.Tokens, utils.Sanitize(e.Name), utils.Sanitize(e.Description), utils.Sanitize(e.Tags))
}

func (e *Events) Save(s *mgo.Session) error {
	e.buildTokenList()
	index := mgo.Index{
		Key:        []string{"tokens"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err := EventCol(s).EnsureIndex(index)
	if err != nil {
		twiit.Log.Error("Failed to Ensure database index ", "error", err)
		return err
	}

	//we are creating a new event
	if e.Id == "" {
		e.Id = bson.NewObjectId()
	}

	_, err = EventCol(s).UpsertId(e.Id, e)
	if err != nil {

	}
	return err
}

func GetEventById(s *mgo.Session, Id string) *Events {

	if bson.IsObjectIdHex(Id) {
		ObjectId := bson.ObjectIdHex(Id)
		return GetEventByObjectId(s, ObjectId)
	} else {
		return new(Events)
	}
}

func GetEventByObjectId(s *mgo.Session, Id bson.ObjectId) *Events {
	u := new(Events)
	err := EventCol(s).FindId(Id).One(u)
	if err != nil {
		twiit.Log.Error("Error on GetEventByObjectId", "error", err)
	}
	return u
}

func GetAllEvents(s *mgo.Session) []*Events {
	events := []*Events{}
	err := EventCol(s).Find(nil).All(&events)
	if err != nil {
		twiit.Log.Error("Error on GetAllEvents", "error", err)
	}
	return events

}

func GetLatestEvents(s *mgo.Session, max int) []*Events {
	events := []*Events{}
	query := bson.M{"date": bson.M{"$gte": time.Now().AddDate(0, 0, -1)}}
	err := EventCol(s).Find(query).Sort("+date").Limit(max).All(&events)
	if err != nil {
		twiit.Log.Error("Error on GetLatestEvents", "error", err)
	}
	return events
}

func FindEvents(s *mgo.Session, query string, limit int) []*Events {
	e := []*Events{}
	var Query bson.M
	//split the query in words
	processedQuery := strings.Fields(query)
	//if more than one word in processedList we iterate over them and  intersect multiple words in query
	if len(processedQuery) > 1 {
		var searches []bson.M
		for _, word := range processedQuery {
			search := bson.M{"$or": []bson.M{
				bson.M{"tokens": &bson.RegEx{Pattern: word, Options: CaseInsensitive}},
			}}
			searches = append(searches, search)
		}
		Query = bson.M{"$and": searches}
		//otherwise
	} else {
		Query = bson.M{"$or": []bson.M{
			bson.M{"tokens": &bson.RegEx{Pattern: query, Options: CaseInsensitive}},
		}}
	}

	err := EventCol(s).Find(Query).Limit(limit).All(&e)
	if err != nil {
		twiit.Log.Error("Error on FindEvents", "error", err)

	}

	return e
}
