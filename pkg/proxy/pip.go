package proxy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var plpRegex *regexp.Regexp
var plpPageRegex *regexp.Regexp

func init() {
	plpRegex = regexp.MustCompile(`<td>(\d+\.\d+\.\d+\.\d+)</td>(?s:.*?)<td>(\d+)</td>(?s:.*?)<td>.*</td>(?s:.*?)<td>.*</td>(?s:.*?)<td>.*</td>(?s:.*?)<td>(.*)</td>`)
	plpPageRegex = regexp.MustCompile(`More Free Proxies.*Fresh-HTTP-Proxy-List-\d+'>\[(\d+)\]</a>.*Updated`)
}

type plp struct{}

func (s *plp) Wait() {
	time.Sleep(10 * time.Second)
}

func (s *plp) Fetch(page int) ([]string, int, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://list.proxylistplus.com/Fresh-HTTP-Proxy-List-%d", page), nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	pageMatch := plpPageRegex.FindSubmatch(data)
	if len(pageMatch) < 1 {
		return nil, 0, errors.New("can not match page")
	}
	pageNum, err := strconv.Atoi(string(pageMatch[1]))
	if err != nil {
		return nil, 0, err
	}

	proxies := plpRegex.FindAllSubmatch(data, -1)
	if len(proxies) < 1 {
		return nil, 0, errors.New("match proxy list: noting found")
	}

	var result []string
	for _, m := range proxies {
		method := "https"
		if strings.Contains(strings.ToLower(string(m[3])), "no") {
			method = "http"
		}
		result = append(result, fmt.Sprintf("%s://%s:%s", method, string(m[1]), string(m[2])))
	}
	return result, pageNum, nil
}
