package app

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
)

func getHTTPStatus(err error) int {
	if err != nil {
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode
		}
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// dung de phan hoi cac REST
func reponseHTTPrequest(w http.ResponseWriter, body []byte, err error) {
	status := getHTTPStatus(err)
	if status != http.StatusOK {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), status)
	} else {
		if len(body) > 0 {
			w.Header().Set(CONTENTTYPE, APPLICATIONJSON)
			w.WriteHeader(http.StatusOK)
		}
		w.Write(body)
	}
}
