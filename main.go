package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
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
	high_quality := flag.Bool("high-quality", false, "Downloads video at a higher quality")
	threads := flag.Int("t", 8, "Set Thread Count for the HTTP client")
	custom_ffmpeg := flag.String("ff", "-c copy", "Custom FFMPEG arguments")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	defer func() {
		fmt.Println("Press Enter key to exit.")
		reader.ReadByte()
	}()

	// Check for ffmpeg
	_, err := exec.LookPath("ffmpeg")

	if err != nil {
		log.Println(err)
		log.Println("FFMPEG not found. Put ffmpeg.exe in the current dir or in the PATH")
		return
	}

	// log.Printf("FFMEPG found in %s", ffpath)

	if len(*ed_url) <= 0 {
		InteractiveMode(username, password, ed_url, high_quality, reader)
	}

	username_trimmed := strings.TrimSpace(*username)
	password_trimmed := strings.TrimSpace(*password)

	// Parse URL
	u, err := url.Parse(*ed_url)

	if err != nil {
		log.Println(err)
		return
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
			log.Println(err)
			return
		}

		user := strings.TrimSpace(doc.Find("#dropdown08").Text())

		if len(user) <= 0 {
			log.Println("Username or password error")
			return
		}

		log.Println("Logged in as " + user)

		resp, rErr := client.Client.Get(*ed_url)

		if rErr != nil {
			log.Println(err)
			return
		}

		doc2, err := goquery.NewDocumentFromResponse(resp)

		if err != nil {
			log.Println(err)
			return
		}

		title := strings.TrimSpace(doc2.Find("#content-wrapper > div > div.col-md-12 > h2").Text())

		videoName = sanitize.Path(title)
	}

	var res WebServicesResponse

	query := url.Values{}

	query.Add("key", "vhjgyu456dCT")
	query.Add("type", "video_paths")
	query.Add("id", videoId)
	query.Add("full", "ZnVsbA==")

	if err := client.GetJson("https://lecturecapture.sliit.lk/webservice.php", &res, &query); err != nil {
		log.Println(err)
		return
	}

	ur, err := url.Parse("https://lecturecapture.sliit.lk/webservice.php")

	if err != nil {
		log.Println(err)
		return
	}

	if *high_quality {
		if len(res.Video_1_720_m3u8) <= 0 {
			log.Println("High quality is not available")
			return
		}
		log.Println("Selected High Quality video")

		ur.Path = path.Join(u.Path, res.Video_1_720_m3u8)
		videoName += "-high-quality"
	} else {
		ur.Path = path.Join(u.Path, res.Video_1_360_m3u8)
	}

	log.Printf("M3U8 Playlist URL : %s", ur.String())
	log.Println("Starting Download...")

	dErr, m3u8_file := DownloadWithHttp(client.Client, ur.String(), videoName, *threads)

	if dErr != nil {
		log.Println(err)
		return
	}

	before := []string{"-y", "-i", m3u8_file}
	mid := strings.Split(*custom_ffmpeg, " ")
	after := []string{videoName + ".mkv"}

	args := append(before, mid...)
	args = append(args, after...)

	log.Println("Merging all parts together using FFMPEG")
	log.Printf("Using Arguments %v", args)

	if err := FFMPEG(args...); err != nil {
		log.Println(err)
		return
	}

	if err := os.RemoveAll(videoName); err != nil {
		log.Println("Unable to remove " + videoName)
	}

	log.Printf("Video %s download complete. Byeee", videoName)

}

func InteractiveMode(username *string, password *string, ed_url *string, high_quality *bool, reader *bufio.Reader) {
	fmt.Print("~~~~~~~~~ Welcome to Dehemi's Eduscope Downloader ~~~~~~~~~\n\n")
	fmt.Println("--- Enter an Eduscope URL")

	for {
		u, _, _ := reader.ReadLine()

		parsed, err := url.Parse(string(u))

		if err != nil {
			goto wrongurl
		}

		if strings.Compare(parsed.Host, "lecturecapture.sliit.lk") != 0 {
			goto wrongurl
		}

		if len(parsed.Query().Get("id")) <= 0 {
			goto wrongurl
		}

		*ed_url = string(u)

		break
	wrongurl:
		fmt.Println("--- Invalid URL, Enter a correct one")
	}

	fmt.Print("\n--- Use username and password? (y,N)\n--- (That way it will fetch the acctual lecture name and rename the video to it)\n")

	if getChoice(reader) {
		fmt.Println("--- Enter Username")

		*username = getInput(reader, "--- Please enter an username")

		fmt.Println("--- Enter Password")

		*password = getInput(reader, "--- Please enter a password")
	}

	fmt.Println("--- Want High Quality Video? (y,N)")

	if getChoice(reader) {
		*high_quality = true
	}
}

func getChoice(reader *bufio.Reader) bool {
	choice, _, _ := reader.ReadLine()
	low_choice := strings.ToLower(string(choice))

	return (strings.Compare(low_choice, "y") == 0)
}

func getInput(reader *bufio.Reader, invalid string) string {
	input := ""

	for {
		uname, _, _ := reader.ReadLine()

		input = strings.TrimSpace(string(uname))

		if len(input) <= 0 {
			fmt.Println(invalid)
		} else {
			break
		}
	}

	return input
}

func FFMPEG(args ...string) error {
	cmd := exec.Command("ffmpeg", args...)

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
