package jobs

import (
	"errors"
	"fmt"
	"io"
	"my_toolbox/library/log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ImageSendRequestToServer struct {
	Job
}

func (usp *ImageSendRequestToServer) Run() {
	// Proxy ayarlarını yapılandırma
	proxyURL, err := url.Parse("http://176.235.207.143:8099")
	if err != nil {
		fmt.Println("Proxy URL hatası:", err)
		return
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	urlInput := "https://prapazar.net/uploads/12270/products/images/4c81a488ec834f9b889b0cc1775e2ff7.jpg"
	client := http.Client{
		Timeout: 45 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// If there are no previous requests, allow redirect by default
			if len(via) == 0 {
				return nil
			}

			// Get the previous URL and current redirect URL
			prevURL := via[len(via)-1].URL
			currURL := req.URL

			// Check if there’s a switch between HTTP and HTTPS (in either direction)
			isHTTPSToggle := (prevURL.Scheme == "http" && currURL.Scheme == "https") ||
				(prevURL.Scheme == "https" && currURL.Scheme == "http")

			// Check if there’s a "www" to non-"www" change or vice versa
			prevHasWWW := strings.HasPrefix(prevURL.Host, "www.")
			currHasWWW := strings.HasPrefix(currURL.Host, "www.")
			isWWWChange := (prevHasWWW && !currHasWWW) || (!prevHasWWW && currHasWWW)

			// Allow the redirect if it's an HTTP/HTTPS change or "www" change
			if isHTTPSToggle || isWWWChange {
				return nil
			}

			// Block any other redirects
			return errors.New("test")
		},
		Transport: transport,
	}

	req, err := http.NewRequest("GET", urlInput, nil)
	if err != nil {
		log.GetLogger().Graylog(true).Error("image request create error", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:132.0) Gecko/20100101 Firefox/131.0")

	response, errImageRequest := client.Do(req)
	if errImageRequest != nil {
		if strings.Contains(errImageRequest.Error(), "connection reset by peer") {
			fmt.Println("Connection reset by peer error encountered.")
		}
		fmt.Println("image response error", errImageRequest)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("image request read error", err)
	}
	bodyString := string(bodyBytes)

	fmt.Println(response.StatusCode, response.Status, bodyString)

}
