package models

import (
	//"labix.org/v2/mgo"
	//"github.com/HorizontDimension/twiit"
	"labix.org/v2/mgo/bson"
	"time"
)

//entry represent a gues entry
type Entry struct {
	EntryTime time.Time
	CardId    int
	Client    bson.ObjectId
}

type GuestList struct {
	Owner   bson.ObjectId
	Guests  []bson.ObjectId
	Entries []Entry //guest entries (set by doorman)
}

func NewGuestlist(owner bson.ObjectId, guest bson.ObjectId) (guestlist *GuestList) {
	guestlist = new(GuestList)
	guestlist.Owner = owner
	if guest != "" {
		guestlist.Guests = []bson.ObjectId{guest}
	}
	return
}

func (g *GuestList) IsOwner(user bson.ObjectId) bool {
	return g.Owner == user
}

func (g *GuestList) AddGuest(guest bson.ObjectId) {
	//add if not exists
	if !g.GuestExists(guest) {
		g.Guests = append(g.Guests, guest)
	}
}

func (g *GuestList) CheckIn(guest bson.ObjectId, cardid int) {
	entry := Entry{EntryTime: time.Now(), CardId: cardid, Client: guest}
	g.Entries = append(g.Entries, entry)
}

func (g *GuestList) RemoveGuest(guest bson.ObjectId) {

	for i := range g.Guests {
		if g.Guests[i] == guest {
			g.Guests = deleteguest(g.Guests, i)
			break
		}
	}
}

func (g *GuestList) GuestExists(guest bson.ObjectId) bool {
	for i := range g.Guests {
		if g.Guests[i] == guest {
			return true
		}
	}
	return false
}

func deleteguest(g []bson.ObjectId, i int) []bson.ObjectId {
	return append(g[:i], g[i+1:]...)
}
