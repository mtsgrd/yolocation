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
		Status  string
	}

	placeResponseType struct {
		Result placeResultType
		Status string
	}

	placeResultType struct {
		Url string
	}
)

// Searches the Google Places API.
func search(query, location string) *searchResponseType {
	s := strings.Split(location, ";")
	lat, lng := s[0], s[1]

	var params url.Values = map[string][]string{
		"location": {fmt.Sprintf("%v,%v", lat, lng)},
		"key":      {googlePlacesApiKey},
		"query":    {query},
		"radius":   {radius},
		"open":     {"true"},
	}

	apiUrl, _ := url.Parse("https://maps.googleapis.com/maps/api/place/textsearch/json")
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
func getPlaceInfo(placeId string) (string, string) {
	var params url.Values = map[string][]string{
		"key":     {googlePlacesApiKey},
		"placeid": {placeId},
	}

	apiUrl, _ := url.Parse("https://maps.googleapis.com/maps/api/place/details/json")
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
	return res.Result.Url, res.Status
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

func sendYoText(username, message string) {
	var msgQuery url.Values = map[string][]string{
		"text": {message},
	}
	msgUrl := url.URL{Scheme: "http", Host: "www.yotext.co",
		RawQuery: msgQuery.Encode()}
	sendYo(username, msgUrl.String())
}

func (c App) Yo(query string) revel.Result {
	revel.INFO.Println("Handling request for:", c.Request.URL.String())
	username := c.Params.Get("username")
	userLocation := c.Params.Get("location")
	response := search(query, userLocation)
	if response.Status != "OK" {
		status := fmt.Sprintf("%v: %v", "error", response.Status)
		sendYoText(username, status)
		return c.RenderText(status)
	}
	for _, result := range response.Results {
		if strings.ToLower(result.Name) != query {
			continue
		}
		mapUrl, status := getPlaceInfo(result.PlaceId)
		if status != "OK" {
			status := fmt.Sprintf("%v: %v", "error", status)
			sendYoText(username, status)
			return c.RenderText(status)
		}
		sendYo(username, mapUrl)
		return c.RenderText(mapUrl)
	}
	notFoundMessage := fmt.Sprintf("No %v found", query)
	sendYoText(username, notFoundMessage)
	return c.RenderText("No search results found.")
}

func init() {
	log.Println("Yo API Token set:", len(yoApiToken) > 0)
	log.Println("Yo API URL set:", len(yoApiUrl) > 0)
	log.Println("Google API key set:", len(googlePlacesApiKey) > 0)
}
