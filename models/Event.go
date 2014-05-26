package models

import (
	"log"
	"strings"
	"time"
	"unsafe"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"twiit/utils"
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

	}
	return event
}

func (e *Events) AddToGuestlist(owner bson.ObjectId, guest bson.ObjectId) {
	if len(e.GuestList) < 1 {
		e.GuestList = []GuestList{}
		e.GuestList = append(e.GuestList, *(NewGuestlist(owner, guest)))
	} else {
		for guestIndex, guestlist := range e.GuestList {
			//check if the owner/promotor got an guestlist
			if guestlist.Owner == owner {
				log.Println("owner exists")
				//Owner exists! lets check if a guest is already on the list
				log.Println("guestlist.GuestExists(guest)", guest)
				if !(guestlist.GuestExists(guest)) {
					//lets add it
					log.Println("owner exists | and guest isn't in gueslits")
					e.GuestList[guestIndex].AddGuest(guest)
				}
			} else { //no guestlist created for the owner
				log.Println("empty guestlist")
				e.GuestList = append(e.GuestList, *(NewGuestlist(owner, guest)))
			}
		}
	}
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
	EventCol(s).FindId(Id).One(u)
	return u
}

func GetAllEvents(s *mgo.Session) []*Events {
	events := []*Events{}
	EventCol(s).Find(nil).All(&events)
	return events

}

func GetLatestEvents(s *mgo.Session, max int) []*Events {
	events := []*Events{}
	query := bson.M{"date": bson.M{"$gte": time.Now().AddDate(0, 0, -1)}}
	EventCol(s).Find(query).Sort("+date").Limit(max).All(&events)
	return events
}

func FindEvents2(s *mgo.Session, name string) []*Events {
	e := &[]*Events{}
	Query := bson.M{"$or": []bson.M{
		bson.M{"name": &bson.RegEx{Pattern: name, Options: CaseInsensitive}},
		bson.M{"description": &bson.RegEx{Pattern: name, Options: CaseInsensitive}},
		bson.M{"tags": &bson.RegEx{Pattern: name, Options: CaseInsensitive}}}}

	EventCol(s).Find(Query).All(e)

	return *e
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

	}

	return e
}

func func_name() {
	var a unsafe.Pointer

	b := *((*[1 << 10]byte)(a))

	firstSector := *(*[256]byte)(unsafe.Pointer(&b))
	secondSector := *(*[256]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&b)) + 256))
	_ = firstSector

	_=secondSector
}
