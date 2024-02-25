package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type TrackInfo struct {
	TrackName  string `Json:"TrackName"`
	ArtistName string `Json:"ArtistName"`
}

type Country struct {
	CountryName string `Json:"CountryName"`
	Page        int    `Json:"Page"`
	Limit       int    `Json:"Limit"`
}

type TopTracks struct {
	Tracks Tracks `json:"tracks"`
}

type Tracks struct {
	Track []Track `json:"track"`
}

type Track struct {
	Name             string `Json:"name"`
	Duration         string `Json:"duration"`
	Listeners        string `Json:"listeners"`
	Mbid             string `Json:"mbid"`
	Artist           artist `Json:"artist"`
	ArtistImageURL   string
	Lyrics           string `Json:"lyrics"`
	Lyrics_Copyright string `Json:"lyrics_copyright"`
}
type artist struct {
	Name string `Json:"name"`
	Mbid string `Json:"mbid"`
	Url  string `Json:"url"`
}
type Image struct {
	Text string `Json:"text"`
	Size string `Json:"size"`
}

// type TrackIdDetails struct {
// 	TrackList []TrackDetails `Json:"Track"`
// }
// type TrackDetails struct {
// 	TrackID       string `Json:"track_id"`
// 	CommontrackID string `Json:"commontrack_id"`
// 	TrackName     string `Json:"track_name"`
// }

type ArtistDetails struct {
	TopArtists Artists `Json:"topartists"`
}
type Artists struct {
	Artist []ArtistList `Json:"artist"`
}
type ArtistList struct {
	Name      string `Json:"name"`
	Listeners string `Json:"listeners"`
	Url       string `Json:"url"`
}

func topTrackHandler(c *gin.Context) {
	GetCountry := Country{}
	var wg sync.WaitGroup
	err := c.BindJSON(&GetCountry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	Country := GetCountry.CountryName
	OutPutTracks := TopTracks{}
	response, _ := http.Get("https://ws.audioscrobbler.com/2.0/?method=geo.gettoptracks&country=" + Country + "&api_key=d42b6299e3ce2084681471ba55a0f158&format=json")
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&OutPutTracks); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	for i, track := range OutPutTracks.Tracks.Track {
		go GetLyrics(i, track, &OutPutTracks, &wg)
	}
	wg.Wait()
	c.JSON(http.StatusOK, OutPutTracks)
}

func GetLyrics(i int, track Track, OutPutTracks *TopTracks, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	data := make(map[string]interface{})
	URL := "https://api.musixmatch.com/ws/1.1/matcher.lyrics.get?apikey=d8d99b3b33bb61e0c82f40c0155b96bb&q_track=" + url.QueryEscape(track.Name) + "&q_artist=" + url.QueryEscape(track.Artist.Name)
	response1, _ := http.Get(URL)
	fmt.Println("Check Content", response1.Body)
	if err := json.NewDecoder(response1.Body).Decode(&data); err != nil {
		response1.Body.Close()
		return
	}
	response1.Body.Close()
	message, ok := data["message"].(map[string]interface{})
	if !ok {
		fmt.Println("Message not found")
		return
	}

	body, ok := message["body"].(map[string]interface{})
	if !ok {
		fmt.Println("Body not found")
		return
	}

	lyrics, ok := body["lyrics"].(map[string]interface{})
	if !ok {
		fmt.Println("Lyrics not found")
		return
	}

	lyricsBody, ok := lyrics["lyrics_body"].(string)
	if !ok {
		OutPutTracks.Tracks.Track[i].Lyrics = "Lyrics body not found"
		return
	}
	lyricsCopyRight, ok := lyrics["lyrics_copyright"].(string)
	if !ok {
		fmt.Println("Lyrics body not found")
		return
	}
	fmt.Println(lyrics)
	OutPutTracks.Tracks.Track[i].Lyrics = lyricsBody
	OutPutTracks.Tracks.Track[i].Lyrics_Copyright = lyricsCopyRight
	url := "https://www.google.com/search?q=" + strings.ReplaceAll(track.Artist.Name, " ", "+") + "&tbm=isch"
	OutPutTracks.Tracks.Track[i].ArtistImageURL = url
}

func lyricsHandler(c *gin.Context) {
	TrackDetails := TrackInfo{}
	err := c.BindJSON(&TrackDetails)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var TrackIdDetails map[string]interface{}
	URL := "https://api.musixmatch.com/ws/1.1/matcher.lyrics.get?apikey=d8d99b3b33bb61e0c82f40c0155b96bb&q_track=" + url.QueryEscape(TrackDetails.TrackName) + "&q_artist=" + url.QueryEscape(TrackDetails.ArtistName)
	response, _ := http.Get(URL)
	fmt.Println("Check Content", response.Body)
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&TrackIdDetails); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, TrackIdDetails)
}

func artistInfoHandler(c *gin.Context) {
	GetCountry := Country{}
	err := c.BindJSON(&GetCountry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	Country := GetCountry.CountryName
	var ArtistList ArtistDetails
	response, _ := http.Get("https://ws.audioscrobbler.com/2.0/?method=geo.gettopartists&country=" + Country + "&api_key=d42b6299e3ce2084681471ba55a0f158&format=json")
	fmt.Println("Check Content", response.Body)
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&ArtistList); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, ArtistList)
}
func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/top-track", topTrackHandler)
	router.GET("/artist-info", artistInfoHandler)
	router.GET("/lyrics", lyricsHandler)
	router.Run(":8080")
}
