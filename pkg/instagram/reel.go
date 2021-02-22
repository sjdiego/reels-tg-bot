package instagram

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reels-tg-bot/pkg/env"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

const (
	userAgent string = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.152 Safari/537.36"
)

// Get func
func Get(code string) string {
	videoURL, status := makeRequest(code)

	if !status {
		fmt.Println(fmt.Sprintf("Error getting video: %v", videoURL))
	}

	videoPath := "/tmp/" + code
	tmpFolder := env.GetEnv("TMP_FOLDER")
	if len(tmpFolder) > 1 {
		videoPath = strings.TrimRight(tmpFolder, "/") + "/" + code
	}

	if !fileExists(videoPath) {
		err := downloadFile(videoURL, videoPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error downloading video: %s", err.Error()))
			videoPath = ""
		}
	} else {
		fmt.Println(fmt.Sprintf("File already exists. Skipping download."))
	}

	return videoPath
}

func makeRequest(code string) (string, bool) {
	httpClient := http.Client{Timeout: 3 * time.Second}
	endpoint := fmt.Sprintf("https://www.instagram.com/reel/%s/?__a=1", code)

	request, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	request.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(request)

	if err != nil {
		return err.Error(), false
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	if response.StatusCode != 200 {
		return response.Status, false
	}

	body, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return readErr.Error(), false
	}

	result := gjson.GetBytes(body, "graphql.shortcode_media.video_url")

	if len(result.Str) > 0 {
		return result.Str, true
	}

	return result.Str, false
}

func downloadFile(url string, videoPath string) error {
	fmt.Println("Trying to download video...")
	out, err := os.Create(videoPath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading file for '%s', status: %d\n", url, resp.StatusCode)
		return err
	}
	fmt.Printf("Download successful, status: %d\n", resp.StatusCode)
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
