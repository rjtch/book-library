package swaggerui

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	sh := http.StripPrefix("/swaggerui/", staticServer)
	router.PathPrefix("/swaggerui/").Handler(sh)
	

	log.Fatal(http.ListenAndServe(":8080", router))
}
