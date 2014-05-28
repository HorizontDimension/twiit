package resources

import (
	"io"
	"net/http"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/HorizontDimension/twiit"
	"github.com/HorizontDimension/twiit/models"
	"github.com/emicklei/go-restful"
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

	if id == "" {
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusBadRequest, "empty file id")
		if err != nil {
			twiit.Log.Error("Error writing response on GetFile ", "error", err)

		}
		return
	}

	var files *mgo.GridFS
	files = models.FilesFs(f.Session)

	file, err := files.OpenId(bson.ObjectIdHex(id))
	defer func() {
		err := file.Close()
		if err != nil {
			twiit.Log.Error("Error Open grifsFile ", "error", err, "file", id)
		}

	}()

	if err != nil {

		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusInternalServerError, "error")
		if err != nil {
			twiit.Log.Error("Error writing response on GetFile ", "error", err)
		}
		return
	}

	//set content Type
	response.AddHeader("Content-Type", file.ContentType())
	_, err = io.Copy(response, file)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		err := response.WriteErrorString(http.StatusInternalServerError, err.Error())
		if err != nil {
			twiit.Log.Error("Error writing response on GetFile ", "error", err)

		}
		return
	}

}
