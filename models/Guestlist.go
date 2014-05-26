package models

import (
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	//"time"
)

type GuestList struct {
	Owner  bson.ObjectId
	Guests []bson.ObjectId
}

func NewGuestlist(owner bson.ObjectId, guest bson.ObjectId) (guestlist *GuestList) {
	guestlist = new(GuestList)
	guestlist.Owner = owner
	guestlist.Guests = []bson.ObjectId{guest}
	return
}

func (g *GuestList) AddGuest(guest bson.ObjectId) {
	g.Guests = append(g.Guests, guest)
}

func (g *GuestList) GuestExists(guest bson.ObjectId) bool {
	for _, g := range g.Guests {
		if g == guest {
			return true
		}
	}
	return false
}
