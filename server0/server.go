package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

var cachestoremap = make(map[int]string)

type respentry struct {
	Key   int    `json:"key"`
	Value string `json:"value"`
}

//PutKey in map
func putKey(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	fmt.Println("putKey called")
	keyid, err := strconv.Atoi(p.ByName("key_id"))

	if err != nil {
		fmt.Printf("%s", err)
		rw.WriteHeader(400)
		return
	}

	value := (p.ByName("value"))

	cachestoremap[keyid] = value

	rw.WriteHeader(200)
}

//Getvalueforkey of server1
func getvalueforkey(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	intid, err := strconv.Atoi(p.ByName("key_id"))
	if err != nil {
		fmt.Printf("%s", err)
		panic(err)
	}
	valuestored, ok := cachestoremap[intid]
	if !ok {
		rw.WriteHeader(400)
		return
	}

	resp := respentry{
		Key:   intid,
		Value: valuestored,
	}
	//marshal struct to json
	repjson, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	rw.WriteHeader(200)
	fmt.Fprintf(rw, "%s", repjson)

}

//GetAllKeys of server1
func getAllKeys(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var entriesSlice = make([]respentry, 0)

	for k, v := range cachestoremap {
		entriesSlice = append(entriesSlice, respentry{k, v})
	}

	//marshal struct to json
	repjson, err := json.Marshal(entriesSlice)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	rw.WriteHeader(200)
	fmt.Fprintf(rw, "%s", repjson)

}

func main() {

	mux := httprouter.New()
	mux.GET("/keys/:key_id", getvalueforkey)
	mux.PUT("/keys/:key_id/:value", putKey)
	mux.GET("/keys", getAllKeys)
	server := http.Server{
		Addr:    "localhost:3000",
		Handler: mux,
	}
	server.ListenAndServe()
}
