package main

import (
	"flag"
	"log"
	"net/url"
	"os/exec"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
)

type WebServicesResponse struct {
	Video_1_m3u8_list string `json:"video_1_m3u8_list"`
	Video_1_360_m3u8  string `json:"video_1_360_m3u8"`
	Video_1_720_m3u8  string `json:"video_1_720_m3u8"`
	Video_2_m3u8_list string `json:"video_2_m3u8_list"`
	Video_2_360_m3u8  string `json:"video_2_360_m3u8"`
	Video_2_720_m3u8  string `json:"video_2_720_m3u8"`
}

func main() {
	ed_url := flag.String("url", "", "Eduscope URL")
	username := flag.String("u", "", "Eduscope User Name")
	password := flag.String("p", "", "Eduscope Password")

	flag.Parse()

	if len(*ed_url) <= 0 {
		log.Println("URL is empty")
		return
	}

	username_trimmed := strings.TrimSpace(*username)
	password_trimmed := strings.TrimSpace(*password)

	// Parse URL
	u, err := url.Parse(*ed_url)

	if err != nil {
		log.Panic(err)
	}

	videoId := strings.TrimSpace(u.Query().Get("id"))

	log.Println(videoId)
	if len(videoId) <= 0 {
		log.Println("This is not a URL to a eduscope video.")
		return
	}

	videoName := sanitize.Path(videoId)
	client := NewHttpClient()

	if len(username_trimmed) > 0 {
		log.Println(username_trimmed, password_trimmed)

		resp, err := client.PostForm("https://lecturecapture.sliit.lk/login.php", url.Values{
			"inputEmail":    {username_trimmed},
			"inputPassword": {password_trimmed},
			"submit":        {""},
		})

		if err != nil {
			log.Println("PostForm Error")
			return
		}

		doc, err := goquery.NewDocumentFromResponse(resp)

		if err != nil {
			log.Panic(err)
		}

		user := strings.TrimSpace(doc.Find("#dropdown08").Text())

		if len(user) <= 0 {
			log.Println("Username or password error")
			return
		}

		log.Println("Logged in as " + user)

		resp, rErr := client.Client.Get(*ed_url)

		if rErr != nil {
			log.Panic(err)
		}

		doc2, err := goquery.NewDocumentFromResponse(resp)

		if err != nil {
			log.Panic(err)
		}

		title := strings.TrimSpace(doc2.Find("#content-wrapper > div > div.col-md-12 > h2").Text())

		videoName = sanitize.Path(title)
	}

	log.Println(videoName)

	var res WebServicesResponse

	query := url.Values{}

	query.Add("key", "vhjgyu456dCT")
	query.Add("type", "video_paths")
	query.Add("id", videoId)
	query.Add("full", "ZnVsbA==")

	if err := client.GetJson("https://lecturecapture.sliit.lk/webservice.php", &res, &query); err != nil {
		log.Panicln(err)
	}

	ur, err := url.Parse("https://lecturecapture.sliit.lk/webservice.php")

	if err != nil {
		log.Panicln(err)
	}

	ur.Path = path.Join(u.Path, res.Video_1_360_m3u8)

	if err := DownloadWithFFMPEG(ur.String(), videoName); err != nil {
		log.Panicln(err)
	}
	// log.Println(client.GetString(""))
}

func DownloadWithFFMPEG(url string, name string) error {
	cmd := exec.Command("ffmpeg", "-i", url, "-c", "copy", name+".mkv")

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
