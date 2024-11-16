package main

import (
	//"encoding/json"
	"net/http"
)

func (a *applicationDependencies) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
//fmt.Fprintln(w, "status: available")
//fmt.Fprintf(w, "environment: %s\n", a.config.environment)
//fmt.Fprintf(w, "version: %s\n", appVersion)

//jsResponse := `{"status": "available", "environment": %q, "version": %q}`
//jsResponse = fmt.Sprintf(jsResponse, a.config.environment, appVersion)
//w.Header().Set("Content-Type", "application/json")
//w.Write([]byte(jsResponse))

//encoding JSON V2
//data := map[string]string{
//"status":      "available",
//"environment": a.config.environment,
//"version":     appVersion,
//}

//jsResponse, err := json.Marshal(data)
//if err != nil {
//	a.logger.Error(err.Error())
//	http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
//	return
//}

//jsResponse = append(jsResponse, '\n')
//w.Header().Set("Content-Type", "application/json")
//w.Write(jsResponse)

//v5

data := envelope {
	"status": "available",
	"system_info": map[string]string{
		"environment": a.config.environment,
		"version": appVersion,
	},
}

err := a.writeJSON(w, http.StatusOK, data, nil)
if err != nil {
//a.logger.Error(err.Error())
//http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
//return
a.serverErrorResponse(w, r, err)
}

}

