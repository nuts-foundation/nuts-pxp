package http

import (
	"fmt"
	"net/http"
)

func ErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	fmt.Println("error: " + err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
