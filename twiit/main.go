package main

import (
	"bytes"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/HorizontDimension/twiit/assets"
	"github.com/HorizontDimension/twiit/resources"

	"github.com/HorizontDimension/go-restful/swagger"
	"github.com/emicklei/go-restful"
	"github.com/fengsp/knight"
	"labix.org/v2/mgo"
)

func main() {
	session, err := mgo.Dial("localhost")
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
	doormanR := resources.Doorman{Session: session}

	promotorR.Register(wsContainer)
	guestsR.Register(wsContainer)
	eventsR.Register(wsContainer)
	filesR.Register(wsContainer)
	authR.Register(wsContainer)
	adminsR.Register(wsContainer)
	doormanR.Register(wsContainer)

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
		SwaggerPath: "/apidocs/",
		//	SwaggerFilePath: "/root/gocode/src/github.com/HorizontDimension/twiit/swagger-ui/dist",
		StaticHandler: &BinaryHandler{},
	}

	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("start listening on localhost:80")
	//server := &http.Server{Addr: ":80", Handler: wsContainer}
	//log.Fatal(server.ListenAndServe())
	knight := knight.NewKnight("../")
	log.Fatalln(knight.ListenAndServe(":80", wsContainer))

}

type BinaryHandler struct {
}

func (b *BinaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	file := strings.TrimPrefix(r.RequestURI, "/apidocs/")
	if file == "" {
		file = "index.html"
	}
	mimetype := mime.TypeByExtension(filepath.Ext(file))
	//file = "dist/" + file

	w.Header().Set("Content-Type", mimetype)

	data, err := assets.Asset(file)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}
	if len(data) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}
	io.Copy(w, bytes.NewBuffer(data))
}
