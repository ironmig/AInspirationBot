package main
import (
	"io"
	"net/http"
	"time"
	"fmt"
	"bytes"
	"encoding/json"
	"encoding/base64"
	"os"
	"github.com/ChimeraCoder/anaconda"
	"errors"
	"net/url"
)
const maxmem int64 = 80000000
const inspirobotapi = "http://inspirobot.me/api"
const twitteruploadapi = "https://upload.twitter.com/1.1/media/upload.json"

var interval time.Duration = 15*time.Minute

func GetPicture () (io.ReadCloser,error) {
	res,err := http.Get(inspirobotapi)
	if err != nil {
		return nil,err
	}
	return res.Body,nil
}

var tweets = 0

func PostTweet (api *anaconda.TwitterApi,reader io.ReadCloser) error {
	buff := bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding,&buff)
	
	_,err := io.Copy(encoder,reader)
	if err != nil {
		return err
	}
	reader.Close()
	encoder.Close()

	
	media,err := api.UploadMedia(buff.String())
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("media_ids",fmt.Sprintf("%d",media.MediaID))
	_,err = api.PostTweet("",values)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
func GrabAndPost (a *anaconda.TwitterApi) error {
	reader,err := GetPicture()
	if err != nil {
		return err
	}
	err = PostTweet(a,reader)
	if err != nil {
		return err
	}
	return nil
}
type AuthInfo struct {
	ConsumerKey string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	Secret string `json:"secret"`
	Token string `json:"token"`
}
func GetAuth () (*anaconda.TwitterApi ,error) {
	file, err := os.Open("auth.json")
	if err != nil {
		return nil,err
	}
	decoder := json.NewDecoder(file)
	info := AuthInfo{}
	err = decoder.Decode(&info)
	if err != nil {
		return nil,err
	}

	anaconda.SetConsumerKey(info.ConsumerKey)
	anaconda.SetConsumerSecret(info.ConsumerSecret)
	api := anaconda.NewTwitterApi(info.Token, info.Secret)
	
	ok,err := api.VerifyCredentials()
	if err != nil {
		return nil,err
	}
	if !ok {
		return nil,errors.New("Invalid Credentials")
	}
	return api,nil
}
func main () {
	api,err := GetAuth()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	//Infinitely post tweets until SIGINT
	for {
		err = GrabAndPost(api)
		if err != nil {
			fmt.Println(err)
			time.Sleep(30*time.Second)
			continue
		}
		fmt.Println("Tweet sent at",time.Now())
		time.Sleep(interval)
	}
}