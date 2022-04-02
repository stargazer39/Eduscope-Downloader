package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
	"github.com/pieterclaerhout/go-waitgroup"
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

	if *high_quality {
		if len(res.Video_1_720_m3u8) <= 0 {
			log.Println("High quality is not available")
		}

		ur.Path = path.Join(u.Path, res.Video_1_720_m3u8)
		videoName += "-high-quality"
	} else {
		ur.Path = path.Join(u.Path, res.Video_1_360_m3u8)
	}

	log.Println(ur.String())

	if err := DownloadWithHttp(client.Client, ur.String(), videoName); err != nil {
		log.Panicln(err)
	}
}

func DownloadWithFFMPEG(url string, name string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", url, "-c", "copy", name+".mkv")

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func DownloadWithHttp(client *http.Client, u string, name string) error {
	resp, err := client.Get(u)

	if err != nil {
		return err
	}

	reader := bufio.NewReader(resp.Body)
	var files []string

	for {
		line, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		}

		if strings.HasSuffix(string(line), ".ts") {
			files = append(files, string(line))
		}
	}

	if _, err := os.Stat(name); os.IsNotExist(err) {
		if err := os.Mkdir(name, 0644); err != nil {
			return err
		}
	}

	m3u8_path := path.Join(name, "out.m3u8")

	if err := DownloadURLToPath(client, u, m3u8_path); err != nil {
		log.Panic(err)
	}

	wg := waitgroup.NewWaitGroup(16)

	for _, f := range files {
		wg.BlockAdd()

		go func(fName string) {
			url, _ := url.Parse(u)

			url.Path = path.Join(url.Path, "../"+fName)

			new_url := url.String()
			new_fPath := path.Join(name, fName)

			if err := DownloadURLToPath(client, new_url, new_fPath); err != nil {
				log.Panic(err)
			}
			wg.Done()
		}(f)

	}

	wg.Wait()

	// Merge all together
	cmd := exec.Command("ffmpeg", "-y", "-i", m3u8_path, "-c", "copy", name+".mkv")

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	// if err := os.RemoveAll(name); err != nil {
	// 	log.Println("unable to remove " + name)
	// }

	return nil
}

func DownloadURLToPath(client *http.Client, url string, output string) error {
	file_resp, err := client.Get(url)

	if err != nil {
		return err
	}

	defer file_resp.Body.Close()

	stat, err := os.Stat(output)

	if err != nil {
		// log.Println(err)
	} else {
		if stat.Size() == file_resp.ContentLength {
			log.Println("File already exist, and Size is match so not downloading. \n" + output)
			return nil
		}
	}

	file, err := os.Create(output)

	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := io.CopyBuffer(file, file_resp.Body, nil); err != nil {
		return err
	}

	return nil
}
