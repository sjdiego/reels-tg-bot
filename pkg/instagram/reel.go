package instagram

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

	downloadPath := "/tmp/" + code

	if !fileExists(downloadPath) {
		err := downloadFile(videoURL, downloadPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error downloading video: %s", err.Error()))
		}
	} else {
		fmt.Println(fmt.Sprintf("File already exists. Skipping download."))
	}

	return downloadPath
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

	return result.Str, true
}

func downloadFile(url string, downloadPath string) error {
	fmt.Println(fmt.Sprintf("Trying to download file from %s", url))
	out, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading file for '%s', status: %d", url, resp.StatusCode)
		return err
	}
	fmt.Printf("Downloaded file for '%s', status: %d\n\n", url, resp.StatusCode)
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
