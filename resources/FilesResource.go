package resources

import (
	"net/http"

	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"

	"github.com/emicklei/go-restful"
	"twiit/models"
)

type File struct {
	Session *mgo.Session
}

func (f *File) Register(container *restful.Container) {

	ws := new(restful.WebService)
	ws.
		Path("/files").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.GET("/{file-id}").To(f.GetFile).
		Doc("get a file").
		Operation("GetUser").
		Param(ws.PathParameter("file-id", "id of the file").DataType("string")).
		Writes("file")) // on the response

	container.Add(ws)

}

func (f *File) GetFile(request *restful.Request, response *restful.Response) {

	id := request.PathParameter("file-id")
	log.Println(id)

	if id == "" {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "empty file id")
		return
	}

	var files *mgo.GridFS
	files = models.FilesFs(f.Session)

	file, err := files.OpenId(bson.ObjectIdHex(id))
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "file not found: "+err.Error())
		return
	}
	defer file.Close()

	//set content Type
	response.AddHeader("Content-Type", file.ContentType())
	_, err = io.Copy(response, file)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

}
