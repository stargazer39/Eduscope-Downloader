package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/pieterclaerhout/go-waitgroup"
)

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

		if err != nil {
			switch err {
			case io.EOF:
				break
			default:
				return err
			}
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
		return err
	}

	wg := waitgroup.NewWaitGroup(16)
	progress := make(chan int)
	count := 0
	total := len(files)
	go func() {
		for {
			<-progress
			count++

			log.Printf("Progress - %d/%d Complete.", count, total)
		}
	}()

	for _, f := range files {
		wg.BlockAdd()

		go func(fName string) {
			defer wg.Done()

			url, _ := url.Parse(u)

			url.Path = path.Join(url.Path, "../"+fName)

			new_url := url.String()
			new_fPath := path.Join(name, fName)

		retry:
			if err := DownloadURLToPath(client, new_url, new_fPath); err != nil {
				log.Printf("File %s faild to download with Error :\n%v\nRetrying...", fName, err)
				time.Sleep(time.Second * 5)

				goto retry
			}

			progress <- 1
		}(f)

	}

	wg.Wait()

	log.Println("Merging all parts together using FFMPEG")
	// Merge all together
	cmd := exec.Command("ffmpeg", "-y", "-i", m3u8_path, "-c", "copy", name+".mkv")

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	if err := os.RemoveAll(name); err != nil {
		log.Println("unable to remove " + name)
	}

	log.Printf("%s download complete. Byeee", name)
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
