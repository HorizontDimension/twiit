package models

import (
	//"fmt"
	"strings"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/HorizontDimension/twiit"
	"github.com/HorizontDimension/twiit/utils"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	UserAdmin        uint8 = 0
	UserPromotor     uint8 = 1
	UserClient       uint8 = 2
	UserGateKeeper   uint8 = 3
	CaseInsensitive        = "i"
	MaxNumberResults       = 20
)

//Kind pode ser Promotor|Convidado|Evento
type SearchResult struct {
	Tokens    []string `bson:"tokens" json:"tokens"`
	Kind      string
	Url       string
	EditUrl   string
	RemoveUrl string
	Image     string
	Value     string `bson:"value" json:"value"`
	IsGuest   bool   `bson:"isguest" json:"isguest"`
	Id        string
}

type User struct {
	Id             bson.ObjectId `bson:"_id,omitempty"`
	Firstname      string
	Lastname       string
	Email          string
	PhoneNumber    string
	Created        time.Time
	Updated        time.Time
	Promotors      []bson.ObjectId `bson:",omitempty"` //one or more promotors associated
	Age            int
	Role           uint8
	Photo          bson.ObjectId `bson:",omitempty"`
	Thumb          bson.ObjectId `bson:",omitempty"`
	HashedPassword []byte
	Tokens         []string

	Pass        string `bson:"-" json:"-"`
	PassConfirm string `bson:"-" json:"-"`
}

//func (user *User) String() string {
//	return fmt.Sprintf("%s %s", user.Firstname, user.Lastname)
//}

//func (user *User) Validate(v *revel.Validation) {
//	v.Required((user).Firstname).Key("user.Firstname")
//	v.Check(user.Lastname,
//		revel.Required{},
//		revel.MinSize{1},
//		revel.MaxSize{64},
//	).Key("user.Lastname").Message("")
//	v.Check(user.Lastname, revel.Required{}).Key("user.Lastname").Message("Ob")
//	v.Check(user.Email,
//		revel.Required{},
//	).Key("user.Email")
//	v.Email(user.Email).Message("invalid email").Key("user.Email")
//}

func (u *User) buildTokenList() {
	u.Tokens = []string{}
	u.Tokens = append(u.Tokens, utils.Sanitize(u.Firstname), utils.Sanitize(u.Lastname), utils.Sanitize(u.Email), utils.Sanitize(u.PhoneNumber))
}

// Save a user to the database. If a struct with p.Pass != nil is passed this
// will update the user's password as well.
// This returns the error value from mgo.Upsert()
func (user *User) Save(s *mgo.Session) error {
	// Calculate the new password hash or load the existing one so we don't clobber it on save.
	user.buildTokenList()
	// Index
	index := mgo.Index{
		Key:        []string{"tokens", "role"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err := UserCol(s).EnsureIndex(index)
	if err != nil {
		twiit.Log.Error("Failed to Ensure database index ", "error", err)
		return err

	}

	//if we provide a password
	if user.Pass != "" {
		//if old password exists we are updating password
		if user.HashedPassword != nil {
			user.Updated = time.Now()
		} else { //we are creating a new user
			user.Created = time.Now()
			user.Updated = user.Created
			//we hash the password
			user.HashedPassword, _ = bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
		}
	} else { //if we dont preovide password
		existing := GetUserByObjectId(s, user.Id)
		if existing.HashedPassword != nil {
			user.Updated = time.Now()
			user.HashedPassword = existing.HashedPassword
		}
	}

	_, err = UserCol(s).Upsert(bson.M{"_id": user.Id}, user)
	if err != nil {
		twiit.Log.Error("Unable to save user account", "user", user, "error", err)
		return err
	}
	return nil
}

func UserCol(s *mgo.Session) *mgo.Collection {
	return s.DB("twiit").C("Users")
}

func GetUserByObjectId(s *mgo.Session, Id bson.ObjectId) *User {
	u := new(User)
	err := UserCol(s).FindId(Id).One(u)
	if err != nil {
		twiit.Log.Error("Error on GetUserByObjectId", "error", err)
	}

	return u
}

func GetUserById(s *mgo.Session, Id string) *User {
	if bson.IsObjectIdHex(Id) {
		ObjectId := bson.ObjectIdHex(Id)
		return GetUserByObjectId(s, ObjectId)
	} else {
		return new(User)
	}
}

func GetUserByPhone(s *mgo.Session, phoneNumber string) *User {
	u := new(User)
	err := UserCol(s).Find(bson.M{"PhoneNumber": phoneNumber}).One(u)
	if err != nil {
		twiit.Log.Error("Error on GetUserByPhone", "error", err)
	}
	return u
}

func GetUserByEmail(s *mgo.Session, email string) *User {
	u := new(User)
	err := UserCol(s).Find(bson.M{"email": email}).One(u)
	if err != nil {
		twiit.Log.Error("Error on GetUserByEmail", "error", err)
	}
	return u
}

func GetAllPromotors(s *mgo.Session) (p []*User) {
	p = []*User{}
	err := UserCol(s).Find(bson.M{"role": UserPromotor}).Select(bson.M{"firstname": 1, "lasttname": 1}).All(&p)
	if err != nil {
		twiit.Log.Error("Error on GetAllPromotors", "error", err)
	}
	return p
}

func (u *User) Delete(s *mgo.Session) error {
	return UserCol(s).RemoveId(u.Id)
}

//finduser
func FindUser(s *mgo.Session, query string, role uint8, limit int) []*User {
	u := &[]*User{}
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
		Query = bson.M{"$and": searches, "role": role}
		//otherwise
	} else {
		Query = bson.M{"$or": []bson.M{
			bson.M{"tokens": &bson.RegEx{Pattern: query, Options: CaseInsensitive}, "role": role},
		}}
	}

	err := UserCol(s).Find(Query).Limit(limit).All(u)
	if err != nil {
		twiit.Log.Error("Error on FinUser", "error", err)
	}

	return *u
}

func FindUsersPaginate(s *mgo.Session, role uint8, query string, page int, documentsPerPage int) (u []*User, totalPages int) {

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
		Query = bson.M{"$and": searches, "role": role}
		//otherwise
	} else {
		Query = bson.M{"$or": []bson.M{
			bson.M{"tokens": &bson.RegEx{Pattern: query, Options: CaseInsensitive}, "role": role},
		}}
	}

	totalDocuments, err := UserCol(s).Find(Query).Count()
	if err != nil {
		twiit.Log.Error("Error counting total documents", "error", err)
		return nil, totalDocuments
	}
	if totalDocuments%documentsPerPage != 0 {
		totalPages = totalDocuments/documentsPerPage + 1
	} else {
		totalPages = totalDocuments / documentsPerPage
	}
	if totalPages < page {
		page = totalPages
	}
	// The number of documents to skip.
	skip := documentsPerPage * (page - 1)

	u = []*User{}

	err = UserCol(s).Find(Query).Limit(documentsPerPage).Skip(skip).All(&u)
	if err != nil {
		twiit.Log.Error("Error counting total documents", "error", err)
		return nil, totalPages
	}

	return u, totalPages
}
