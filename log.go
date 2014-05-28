package twiit

import (
	log "gopkg.in/inconshreveable/log15.v1"
)

var errorlogfile = "errorlog.txt"
var errorlog = "log.txt"

var Log = log.New()

func init() {

	handler := log.MultiHandler(
		log.LvlFilterHandler(log.LvlError, log.Must.FileHandler(errorlogfile, log.LogfmtFormat())),
	)

	Log.SetHandler(log.SyncHandler(handler))
}
