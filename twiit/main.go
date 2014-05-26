package main

import (
	"log"
	"net/http"

	"twiit/resources"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"labix.org/v2/mgo"
)

func main() {
	session, err := mgo.Dial("192.168.1.104")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	wsContainer := restful.NewContainer()
	guestsR := resources.Guest{Session: session}
	eventsR := resources.Event{Session: session}
	filesR := resources.File{Session: session}
	authR := resources.Auth{Session: session}
	promotorR := resources.Promotor{Session: session}
	adminsR := resources.Admin{Session: session}

	promotorR.Register(wsContainer)
	guestsR.Register(wsContainer)
	eventsR.Register(wsContainer)
	filesR.Register(wsContainer)
	authR.Register(wsContainer)
	adminsR.Register(wsContainer)

	//wsContainer.Filter(wsContainer.OPTIONSFilter)
	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"TwiitAPI"},
		AllowedHeaders: []string{"Content-Type", "Accept", "Authorization"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)

	// Add container filter to respond to OPTIONS
	wsContainer.Filter(wsContainer.OPTIONSFilter)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "http://guestlist.twiit.pt",
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "/root/gocode/src/twiit/swagger-ui/dist"}
	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("start listening on localhost:80")
	server := &http.Server{Addr: ":8080", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
