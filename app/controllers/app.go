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

var yoApiUrl = os.Getenv("YO_API_URL")
var yoApiToken = os.Getenv("YO_API_TOKEN")
var googlePlacesApiKey = os.Getenv("GOOGLE_API_KEY")

const radius = "50000"

type locationType struct {
	Lat float64
	Lng float64
}

type searchResultsType []struct {
	Name    string
	PlaceId string `json:"place_id"`
}

type searchResponseType struct {
	Results searchResultsType
}

type placeResponseType struct {
	Result placeResultType
}

type placeResultType struct {
	Url string
}

func search(query, location string) *searchResponseType {
	apiUrl, _ := url.Parse("https://maps.googleapis.com/maps/api/place/textsearch/json")
	params := apiUrl.Query()
	s := strings.Split(location, ";")
	lat, lng := s[0], s[1]
	params.Add("location", fmt.Sprintf("%v,%v", lat, lng))
	params.Add("key", googlePlacesApiKey)
	params.Add("query", query)
	params.Add("radius", radius)
	params.Add("open", "true")
	apiUrl.RawQuery = params.Encode()

	log.Println("Executing search :", apiUrl.String())
	resp, err := http.Get(apiUrl.String())
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("Request failed:", apiUrl)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var res searchResponseType

	err = json.Unmarshal(body, &res)

	if err != nil {
		log.Fatal("Invalid json response.", apiUrl)
		panic(err)
	}
	return &res
}

func getMapUrl(placeId string) string {
	apiUrl, _ := url.Parse("https://maps.googleapis.com/maps/api/place/details/json")
	params := apiUrl.Query()
	params.Add("key", googlePlacesApiKey)
	params.Add("placeid", placeId)
	apiUrl.RawQuery = params.Encode()

	log.Println("Requesting place details:", apiUrl.String())
	resp, err := http.Get(apiUrl.String())
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("Request failed:", apiUrl)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var res placeResponseType

	err = json.Unmarshal(body, &res)

	if err != nil {
		log.Fatal("Invalid json response.", apiUrl)
		panic(err)
	}
	return res.Result.Url
}

func sendYo(username, link string) {
	log.Println("Sending yo link to:", username, link)
	_, err := http.PostForm(yoApiUrl, url.Values{"api_token": {yoApiToken},
		"username": {username}, "link": {link}})
	if err != nil {
		log.Fatal("Failed to send Yo.")
		panic(err)
	}
}

func (c App) Yo(query string) revel.Result {
	log.Println("Yo API Token:", yoApiToken)
	log.Println("Yo API URL:", yoApiUrl)
	log.Println("Google API key:", googlePlacesApiKey)
	log.Println(c.Request.URL.String())
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
	return c.RenderText("")
}
