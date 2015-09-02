package app

import (
	"net/http"
)

func init() {

	http.HandleFunc("/_ah/testTQ", testTQ)

}
