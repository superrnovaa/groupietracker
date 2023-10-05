package main

import (
"log"
"net/http"
"text/template"
"strconv"
"strings"
"encoding/json"
"io/ioutil"
//"groupietracker/api"
"os"
"fmt"
)


var artists = []Artist{}
var relation = Relation{}
var allGeodata Geodata
var location  map[string]interface{}
type SearchRequest struct {
	Query string `json:"q"`
}



func syncData(api string, data interface{}) {
	log.Println("Started synchronization api " + api)
	res, err := http.Get(api)
	if err != nil {
		log.Println(err.Error())
		return
	}

		bodyBytes, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Println(err.Error())
			return
		}

		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}



func main() {

	go syncData("https://groupietrackers.herokuapp.com/api/artists", &artists)
	go syncData("https://groupietrackers.herokuapp.com/api/relation", &relation)
	//start()

	router := http.NewServeMux()
	port := ":4000"
	router.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	router.HandleFunc("/artist", artistsHandler)
	router.HandleFunc("/geo",  GetGeocode)
	router.HandleFunc("/search",  searchHandler)
	router.HandleFunc("/", index)
	log.Printf("Listening %s\n", "http://localhost:4000")
	log.Fatal(http.ListenAndServe(port, router))
}

func index(w http.ResponseWriter, r *http.Request) {
	indexTpl := template.Must(template.ParseGlob("index.html"))
	
	 if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		indexTpl.ExecuteTemplate(w, "index.html", artists)
	default:
		http.Error(w, "400 Bad Request: This resource only supports the POST method", http.StatusBadRequest)
	}
}

func artistsHandler(w http.ResponseWriter, r *http.Request) {
	indexTpl := template.Must(template.ParseGlob("artist.html"))
	
	// if r.URL.Path != "/artistsHandler" {
	//	http.Error(w, "404 Not Found", http.StatusNotFound)
	//	return
	//}
	keys, ok := r.URL.Query()["ID"]
	if !ok || len(keys) != 1 {
		log.Println("Url Param 'ID' is missing")
	}
	key := keys[0]
	id, err := strconv.Atoi(key)
	if err != nil {
		log.Println("Url Param 'ID' is missing")
	}

	artist := filterArtistByID(artists, id)
	if artist == nil {
		log.Println("Artist Not Found")
	}

	var data = artistData{Artist: *artist}
	var dates = filterRelationByID(relation.Index, id)

	if dates != nil {
		data.DatesLocations = make(map[string]interface{})
		for key, _ := range dates.DatesLocations {
			var locationName = strings.ReplaceAll(key, "_", " ")
			locationName = strings.ReplaceAll(locationName, "-", " - ")
			locationName = strings.ToUpper(locationName)
			data.DatesLocations[key] = locationName

		}
		location = data.DatesLocations
	}
	//fmt.Println(data.DatesLocations)
	fmt.Println(location)

	switch r.Method {
	case "GET":
		indexTpl.ExecuteTemplate(w, "artist.html", data)
	default:
		http.Error(w, "400 Bad Request: This resource only supports the POST method", http.StatusBadRequest)
		break
	}
}


func GetGeocode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		jsonFile, err := os.Open("geocode.json")
		if err != nil {
			log.Println("Error:", err)
		}
		jsonData, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			log.Println(err)
		}
		defer jsonFile.Close()
		json.Unmarshal(jsonData, &allGeodata)
		fmt.Println(location)
		//cityName := "amsterdam-netherlands"
		var newmap = make(map[string]interface{})
		for key, _ := range location {
			fmt.Println(key)
	// Loop through the maps to find the city name and save the coordinates
	for _, entry := range allGeodata.Index {
		if coords, ok := entry.CountryCoords[key]; ok {
			newmap[key] = coords
			break
		}
	}
		}
		
		b, err := json.Marshal(newmap)
		fmt.Println(newmap)
		if err != nil {
			log.Println("Error during json marshlling. Error:", err)
		}
		w.Write(b)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("This function does not support " + r.Method + " method."))
	}
}


func filterArtistByID(artists []Artist, id int) *Artist {
	for _, item := range artists {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

func filterRelationByID(relations []artistRelation, id int) *artistRelation {
	for _, item := range relations {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Assume artists is a slice of Artist structs containing the available artists
fmt.Println("inside search function")
	// Parse the JSON request body
	var searchReq SearchRequest
	err := json.NewDecoder(r.Body).Decode(&searchReq)
	fmt.Println("nnn")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error parsing JSON request body")
		return
	}
	fmt.Println("ooo")
	fmt.Println(searchReq)
	// Get the search term from the request body
	searchTerm := searchReq.Query
	fmt.Println(searchTerm)
	// Perform the search based on the search term
	semiartists := []Artist{}
	searchingFor := strings.ToLower(searchTerm)

	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), searchingFor) ||
			strings.Contains(strconv.Itoa(artist.CreationDate), searchTerm) ||
			strings.Contains(strings.ToLower(artist.FirstAlbum), searchTerm) {
			semiartists = append(semiartists, artist)
		} else {
			for _, member := range artist.Members {
				if strings.Contains(strings.ToLower(member), searchingFor) {
					semiartists = append(semiartists, artist)
					break
				}
			}
		}
	}
	for _, artist := range semiartists {
		fmt.Println(artist.Name)
	}
	
	// Render the response template with the search results
	indexTpl := template.Must(template.ParseGlob("index.html"))
	err = indexTpl.ExecuteTemplate(w, "index.html", semiartists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error rendering template")
		return
	}
}


