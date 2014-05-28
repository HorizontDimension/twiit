package models

import (
	//"io"
	"github.com/nfnt/resize"

	"github.com/HorizontDimension/twiit"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"mime/multipart"
	"path/filepath"
	"strings"
)

func FilesFs(s *mgo.Session) *mgo.GridFS {
	s.SetMode(mgo.Monotonic, true)
	return s.DB("twiit").GridFS("fs")
}

type ImageFormats struct {
	Width  uint
	Height uint
}

func getFileInfo(filename string) (name string, ext string) {
	splitedFilename := strings.Split(filename, ".")
	ext = filepath.Ext(filename)
	name = strings.Join(splitedFilename[0:len(splitedFilename)-1], ".")
	return
}

//Session must be closed outside this func
func AddImage(s *mgo.Session, name string, fhs []*multipart.FileHeader, fileId bson.ObjectId, format ImageFormats) error {

	//iterate over multiapartFileHeader
	for i := 0; i < len(fhs); i++ {
		f, err := fhs[i].Open()
		if err != nil {
			twiit.Log.Error("Failed to open multipart", "error", err)
			return err
		}

		img, _, err := image.Decode(f)
		if err != nil {
			twiit.Log.Error("Failed to decode image", "error", err)
			return err
		}

		gridfile, err := FilesFs(s).Create("")
		if err != nil {
			twiit.Log.Error("Failed to create gridfile", "error", err)
			return err
		}

		gridfile.SetId(fileId)
		gridfile.SetContentType(fhs[0].Header.Get("Content-Type"))

		m := resize.Resize(format.Width, format.Height, img, resize.NearestNeighbor)

		err = jpeg.Encode(gridfile, m, nil)

		//please handle error ws notification
		if err != nil {
			twiit.Log.Error("Failed to encode image", "error", err)

			return err
		}

		err = gridfile.Close()
		if err != nil {
			twiit.Log.Error("Failed to close gridfile", "error", err)
			return err
		}

		err = f.Close()
		if err != nil {
			twiit.Log.Error("Failed to close multipart file", "error", err)
			return err
		}

	}
	return nil

}
