package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type App struct {
	*revel.Controller
}

// Google Places search radius in meters.
const radius = "50000"

var (
	yoApiUrl           = os.Getenv("YO_API_URL")
	yoApiToken         = os.Getenv("YO_API_TOKEN")
	googlePlacesApiKey = os.Getenv("GOOGLE_API_KEY")
)

// Data structures for parsing JSON responses.
type (
	locationType struct {
		Lat float64
		Lng float64
	}

	searchResultsType []struct {
		Name    string
		PlaceId string `json:"place_id"`
	}

	searchResponseType struct {
		Results searchResultsType
	}

	placeResponseType struct {
		Result placeResultType
	}

	placeResultType struct {
		Url string
	}
)

// Searches the Google Places API.
func search(query, location string) *searchResponseType {

	apiUrl, _ := url.Parse("https://maps.googleapis.com/maps/api/place/textsearch/json")

	// Add query parameters.
	params := apiUrl.Query()
	s := strings.Split(location, ";")
	lat, lng := s[0], s[1]
	params.Add("location", fmt.Sprintf("%v,%v", lat, lng))
	params.Add("key", googlePlacesApiKey)
	params.Add("query", query)
	params.Add("radius", radius)
	params.Add("open", "true")
	apiUrl.RawQuery = params.Encode()

	revel.INFO.Println("Executing search :", apiUrl.String())
	resp, err := http.Get(apiUrl.String())
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("Request failed:", apiUrl)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var res searchResponseType

	if err = json.Unmarshal(body, &res); err != nil {
		log.Fatal("Invalid json response.", apiUrl)
	}
	return &res
}

// Looks up place information and returns a direct link.
func getMapUrl(placeId string) string {
	apiUrl, _ := url.Parse("https://maps.googleapis.com/maps/api/place/details/json")
	params := apiUrl.Query()
	params.Add("key", googlePlacesApiKey)
	params.Add("placeid", placeId)
	apiUrl.RawQuery = params.Encode()

	revel.INFO.Println("Requesting place details:", apiUrl.String())
	resp, err := http.Get(apiUrl.String())
	if err != nil {
		log.Fatal("Request failed:", apiUrl)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var res placeResponseType

	if err = json.Unmarshal(body, &res); err != nil {
		log.Fatal("Invalid json response.", apiUrl)
	}
	return res.Result.Url
}

// Sends a Yo back to the user who invoked the search.
func sendYo(username, link string) {
	revel.INFO.Println("Sending yo link to:", username, link)
	_, err := http.PostForm(yoApiUrl, url.Values{"api_token": {yoApiToken},
		"username": {username}, "link": {link}})
	if err != nil {
		log.Fatal("Failed to send Yo.")
	}
}

func (c App) Yo(query string) revel.Result {
	revel.INFO.Println("Handling request for:", c.Request.URL.String())
	username := c.Params.Get("username")
	userLocation := c.Params.Get("location")
	response := search(query, userLocation)
	for _, result := range response.Results {
		if strings.ToLower(result.Name) != query {
			continue
		}
		mapUrl := getMapUrl(result.PlaceId)
		sendYo(username, mapUrl)
		return c.RenderText(mapUrl)
	}
	if len(response.Results) > 0 {
		mapUrl := getMapUrl(response.Results[0].PlaceId)
		sendYo(username, mapUrl)
		return c.RenderText(mapUrl)
	} else {
		sendYo(username, "")
		return c.RenderText("No search results found.")
	}
}

func init() {
	log.Println("Yo API Token set:", len(yoApiToken) > 0)
	log.Println("Yo API URL set:", len(yoApiUrl) > 0)
	log.Println("Google API key set:", len(googlePlacesApiKey) > 0)
}
