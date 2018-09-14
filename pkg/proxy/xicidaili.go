package proxy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var xicidailiRegex *regexp.Regexp
var xicidailiPageRegex *regexp.Regexp

func init() {
	xicidailiRegex = regexp.MustCompile(`<td>(\d+\.\d+\.\d+\.\d+)</td>(?s:.*?)<td>(\d+)</td>(?s:.*?)<td>(HTTPS?)</td>`)
	xicidailiPageRegex = regexp.MustCompile(`.*>(\d+)</a>.*下一页`)
}

type xicidaili struct {
	baseUrl string
}

func (s *xicidaili) Wait() {}

func (s *xicidaili) Fetch(page int) ([]string, int, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", s.baseUrl, page), nil)
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

	pageMatch := xicidailiPageRegex.FindSubmatch(data)
	if len(pageMatch) < 1 {
		return nil, 0, errors.New("can not match page")
	}
	pageNum, err := strconv.Atoi(string(pageMatch[1]))
	if err != nil {
		return nil, 0, err
	}

	proxies := xicidailiRegex.FindAllSubmatch(data, -1)
	if len(proxies) < 1 {
		return nil, 0, errors.New("match proxy list: noting found")
	}

	var result []string
	for _, m := range proxies {
		result = append(result, fmt.Sprintf("%s://%s:%s", strings.ToLower(string(m[3])), string(m[1]), string(m[2])))
	}
	return result, pageNum, nil
}
