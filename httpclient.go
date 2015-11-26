package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type entries []struct {
	Key   int    `json:"key"`
	Value string `json:"value"`
}

type respentry struct {
	Key   int    `json:"key"`
	Value string `json:"value"`
}

//ErrNodeNotFound
var ErrNodeNotFound = errors.New("node not found")

//Circle ..
type Circle struct {
	Nodes Nodes
}

//NewCircle ..
func NewCircle() *Circle {
	return &Circle{Nodes: Nodes{}}
}

//AddNode ..
func (r *Circle) AddNode(url string) {

	node := NewNode(url)
	r.Nodes = append(r.Nodes, node)

	sort.Sort(r.Nodes)
}

//Get ..it returns the server node where the the passed key should be stored
func (r *Circle) Get(url string) string {
	i := search(r, url)
	if i >= r.Nodes.Len() {
		i = 0
	}

	return r.Nodes[i].URL
}

// func which searches for a node
func search(r *Circle, url string) int {

	return sort.Search(r.Nodes.Len(), func(i int) bool {
		return r.Nodes[i].HashURL >= hashURL(url)
	})
}

//----------------------------------------------------------
// Node
//----------------------------------------------------------

//Node ...
type Node struct {
	//URL ..
	URL string
	//HashURL ..
	HashURL uint32
}

//NewNode ...
func NewNode(url string) *Node {
	return &Node{
		URL:     url,
		HashURL: hashURL(url),
	}
}

//Nodes ..
type Nodes []*Node

func (n Nodes) Len() int           { return len(n) }
func (n Nodes) Swap(x, y int)      { n[x], n[y] = n[y], n[x] }
func (n Nodes) Less(x, y int) bool { return n[x].HashURL < n[y].HashURL }

//----------------------------------------------------------
// Helpers
//----------------------------------------------------------

func hashURL(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

//initalize circle
var c = &Circle{Nodes: Nodes{}}

var (
	node1url = "http://localhost:3000"
	node2url = "http://localhost:3001"
	node3url = "http://localhost:3002"
)

func addNodes() {
	c.AddNode(node1url)
	c.AddNode(node2url)
	c.AddNode(node3url)
}

func putInCache(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	keyid, _ := strconv.Atoi(p.ByName("key_id"))
	value := (p.ByName("value"))

	var buffer bytes.Buffer
	// generate request the uber api	postUberReply
	client := &http.Client{}

	url := c.Get(strconv.Itoa(keyid))

	buffer.WriteString(url)
	buffer.WriteString("/keys/")
	buffer.WriteString(strconv.Itoa(keyid))
	buffer.WriteString("/")
	buffer.WriteString(value)

	r, _ := http.NewRequest("PUT", buffer.String(), nil)
	resp, err := client.Do(r)

	if err != nil {
		fmt.Printf("error occured")
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	if resp.StatusCode == 200 {
		rw.WriteHeader(200)
	} else {
		rw.WriteHeader(500)
	}

}

func getFromCache(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	keyid, err := strconv.Atoi(p.ByName("key_id"))
	if err != nil {
		fmt.Printf("%s", err)
		panic(err)
	}

	var buffer bytes.Buffer
	url := c.Get(strconv.Itoa(keyid))

	buffer.WriteString(url)
	buffer.WriteString("/keys/")
	buffer.WriteString(strconv.Itoa(keyid))

	var s respentry
	response, err := http.Get(buffer.String())
	if err != nil {
		fmt.Printf("error occured")
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		if response.StatusCode == 200 {
			json.Unmarshal([]byte(contents), &s)
			resp := respentry{
				Key:   s.Key,
				Value: s.Value,
			}
			repjson, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("%s", err)
				panic(err)
			}
			rw.WriteHeader(200)
			fmt.Fprintf(rw, "%s", repjson)

		} else {
			rw.WriteHeader(500)
		}
	}
}

func main() {
	addNodes()
	mux := httprouter.New()
	mux.GET("/keys/:key_id", getFromCache)
	mux.PUT("/keys/:key_id/:value", putInCache)
	server := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
