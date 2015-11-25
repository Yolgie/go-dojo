package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

import "flag"

const FORECAST_URL string = "https://api.forecast.io/forecast/%s/%v,%v"
const FORECAST_TOKEN string = "GET YOUR OWN :)"

const OPENCAGE_URL string = "https://api.opencagedata.com/geocode/v1/json"
const OPENCAGE_TOKEN string = "GET YOUR OWN :)"

type coordinate struct {
	longitude float64
	latitude  float64
	name      string
}

type weather struct {
	summary     string
	temperature float64
}

func (coord coordinate) String() string {
	return fmt.Sprintf("Name: %s (%v, %v)", coord.name, coord.longitude, coord.latitude)
}

func (w weather) String() string {
	return fmt.Sprintf("%s mit %v° stupid Farenheit", w.summary, w.temperature)
}

type GeocodingResponse struct {
	Results []OpenCageResult `json:"results"`
}

type OpenCageResult struct {
	Annotations struct {
		Timezone struct {
			Name string `json:"name"`
		}
	} `json:"annotations"`
	Name     string `json:"formatted"`
	Geometry struct {
		Longitude float64 `json:"lng"`
		Latitude  float64 `json:"lat"`
	} `json:"geometry"`
}

type ForecastDatapoint struct {
	Summary     string  `json:"summary"`
	Temperature float64 `json:"temperature"`
}

type ForecastResponse struct {
	Timezone  string            `json:"timezone"`
	Currently ForecastDatapoint `json:"currently"`
}

func getCityCoordinates(city string) (*coordinate, error) {
	// ?q=PLACENAME&key=YOURKEY
	url := OPENCAGE_URL + "?q=" + city + "&key=" + OPENCAGE_TOKEN
	fmt.Println("request " + url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	//io.Copy(os.Stdout, resp.Body)

	result := GeocodingResponse{}
	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rawData, &result)
	if err != nil {
		return nil, err
	}

	return &coordinate{
		longitude: result.Results[0].Geometry.Longitude,
		latitude:  result.Results[0].Geometry.Latitude,
		name:      result.Results[0].Name,
	}, nil
}

func getWeather(coord *coordinate) (*weather, error) {
	url := fmt.Sprintf(FORECAST_URL, FORECAST_TOKEN, coord.latitude, coord.longitude)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	result := ForecastResponse{}
	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rawData, &result)
	if err != nil {
		return nil, err
	}

	return &weather{
		summary:     result.Currently.Summary,
		temperature: result.Currently.Temperature,
	}, nil
}

func main() {
	flag.Parse()
	args := flag.Args()
	weatherChannel := make(chan string, len(args))
	//  weatherData
	for _, city := range args {
		go func(c string) {
			coord, err := getCityCoordinates(c)
			if err != nil {
				log.Fatal(err)
			}
			theWeather, err := getWeather(coord)
			weatherChannel <- theWeather.String()
		}(city)
	}
	idx := 0
	for currentWeather := range weatherChannel {
		fmt.Println(currentWeather)
		idx++
		if idx == len(args) {
			break
		}
	}
}
