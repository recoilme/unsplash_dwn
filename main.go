package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var AccessKey string
var PageOffset uint64
var Query string
var ImgQuery string
var Prefix string
var Captions bool

type Photo struct {
	ID             string     `json:"id,omitempty"`
	AltDescription string     `json:"alt_description,omitempty"`
	Urls           Urls       `json:"urls,omitempty"`
	Links          PhotoLinks `json:"links,omitempty"`
	Exif           Exif       `json:"exif,omitempty"`
	Tags           []struct {
		Title string `json:"title"`
	} `json:"tags,omitempty"`
}

type Urls struct {
	Raw     string `json:"raw,omitempty"`
	Full    string `json:"full,omitempty"`
	Regular string `json:"regular,omitempty"`
	Small   string `json:"small,omitempty"`
	Thumb   string `json:"thumb,omitempty"`
	SmallS3 string `json:"small_s3,omitempty"`
}

type Exif struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	Name         string `json:"name"`
	ExposureTime string `json:"exposure_time"`
	Aperture     string `json:"aperture"`
	FocalLength  string `json:"focal_length"`
	ISO          int64  `json:"iso"`
}

type PhotoLinks struct {
	Self             string `json:"self,omitempty"`
	HTML             string `json:"html,omitempty"`
	Download         string `json:"download,omitempty"`
	DownloadLocation string `json:"download_location,omitempty"`
}

type Photos []Photo

type SearchResults struct {
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
	Results    []struct {
		ID             string `json:"id"`
		AltDescription string `json:"alt_description,omitempty"`
		Urls           Urls   `json:"urls,omitempty"`
	} `json:"results"`
}

func main() {
	flag.StringVar(&AccessKey, "c", "", "Client Access Key")
	flag.StringVar(&Query, "q", "photos", "query string")
	flag.StringVar(&Prefix, "pr", "", "caption prefix")
	flag.BoolVar(&Captions, "captions", false, "create captions")
	flag.StringVar(&ImgQuery, "iq", "&w=256&h=256&fit=crop&crop=faces", "image query")
	flag.Parse()

	if AccessKey == "" {
		log.Panicln("Missing access key.")
	}

	os.MkdirAll("img", os.ModePerm)

	for {

		current := time.Now()
		var photos Photos

		// Crawl Topics photos
		photos, err := getPhotos(AccessKey,
			Query,
			int(PageOffset))

		if err != nil {
			time.Sleep(30 * time.Second)
			continue
		}

		if len(photos) == 0 {
			// Next topics
			PageOffset = 1

			log.Println("Next page: ", PageOffset)
			time.Sleep(72 * time.Second)
			continue
		}
		hasErr := false
		for _, photo := range photos {

			err := downloadFile(photo.Urls.Raw, strings.ReplaceAll(photo.ID, "-", "_"), ImgQuery, Prefix, photo.AltDescription)
			if !hasErr && err != nil {
				hasErr = true
			}
		}

		elapse := int32(time.Since(current).Seconds())

		log.Println("Time used: ", elapse, " seconds.")

		// Next page query only for all downloads success.
		if !hasErr {
			PageOffset += 1
			log.Println("Next page: ", PageOffset)
		}

		diff := 72 - elapse
		offset := int32(0)
		if diff <= 0 {
			offset = 72
		} else if diff > 72 {
			offset = 72
		} else {
			offset = 0
		}
		delay := offset + rand.Int31n(10)
		log.Println("Sleep: ", delay, " seconds.")
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func getPhotos(key string, query string, page int) (Photos, error) {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	likedApi := "https://api.unsplash.com/" + query

	req, _ := http.NewRequest(http.MethodGet, likedApi, nil)

	req.Header.Add("Authorization", "Client-ID "+key)
	q := req.URL.Query()
	q.Add("per_page", "30")
	q.Add("page", strconv.Itoa(page))
	q.Add("order_by", "oldest")
	req.URL.RawQuery = q.Encode()

	log.Println(req.URL.String(), " , query: ", req.URL.Query())

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	//log.Println("Response: ", res)

	if res.StatusCode == 200 {
		// Parse links
		body, _ := io.ReadAll(res.Body)
		//log.Printf("%#v", string(body))
		photos, err := UnmarshalPhotos(body)
		log.Println("photos: ", len(photos), "err: ", err)

		if err != nil {
			var sr SearchResults
			sr_err := json.Unmarshal(body, &sr)
			if sr_err == nil && len(sr.Results) > 0 {
				for _, res := range sr.Results {
					photos = append(photos, Photo{ID: res.ID, AltDescription: res.AltDescription, Urls: res.Urls})
				}
				//log.Println(photos)
				return photos, nil
			}

			return nil, err
		}

		return photos, nil
	} else {
		err = errors.New(res.Status)
		log.Println("Err: ", err)
		return nil, err
	}
}

func UnmarshalPhotos(data []byte) (Photos, error) {
	var r Photos
	err := json.Unmarshal(data, &r)
	return r, err
}

func downloadFile(URL, fileName, iq, prefix, description string) error {
	name := fileName + ".png"
	_, err := os.Stat("img/" + name)

	if err == nil {
		// File exist
		log.Println("Photo ", name, " already downloaded.")
		return nil
	}

	//downloadTokens <- struct{}{}
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	//Get the response bytes from the url
	response, err := client.Get(URL + iq + "&fm=png")
	//log.Println("photo url", URL+iq+"&fm=png")
	if err != nil {
		log.Println(fileName, " request error: ", err)
		return err
	}
	defer response.Body.Close()
	//<-downloadTokens

	if response.StatusCode != 200 {
		log.Println(fileName, " Received non 200 response code: ",
			response.StatusCode)
		return errors.New("received non 200 response code")
	}

	log.Println("Downloaded: ", fileName)
	//Create a empty file
	file, err := os.Create("img/" + name)
	if err != nil {
		log.Println("Fail create file ", fileName)
		return err
	}
	defer file.Close()

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Println("Fail write file ", fileName)
		return err
	}

	if Captions {
		desName := strings.Replace(name, ".png", ".caption", -1)
		fileDesc, err := os.Create("img/" + desName)
		if err != nil {
			log.Println("Fail create file ", fileName)
			return err
		}
		defer fileDesc.Close()

		_, err = io.Copy(fileDesc, strings.NewReader(prefix+description+","))
		if err != nil {
			log.Println("Fail write file desc ", fileName)
			return err
		}
	}

	return nil
}
