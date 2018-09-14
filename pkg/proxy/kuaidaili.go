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

var kuaiDailiRegex *regexp.Regexp
var kuaiDailiPageRegex *regexp.Regexp

func init() {
	kuaiDailiRegex = regexp.MustCompile(`data-title="IP">(.*?)</td>(?s:.*?)data-title="PORT">(.*?)</td>(?s:.*?)data-title="类型">(.*?)</td>`)
	kuaiDailiPageRegex = regexp.MustCompile(`>(\d+)</a></li><li>页</li></ul>`)
}

type kuaidaili struct {
	baseUrl string
}

func (s *kuaidaili) Wait() {}
func (s *kuaidaili) Fetch(page int) ([]string, int, error) {
	client := http.DefaultClient
	resp, err := client.Get(fmt.Sprintf("%s%d/", s.baseUrl, page))
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

	pageMatch := kuaiDailiPageRegex.FindSubmatch(data)
	if len(pageMatch) < 1 {
		return nil, 0, errors.New("can not match page")
	}
	pageNum, err := strconv.Atoi(string(pageMatch[1]))
	if err != nil {
		return nil, 0, err
	}

	proxies := kuaiDailiRegex.FindAllSubmatch(data, -1)
	if len(proxies) < 1 {
		return nil, 0, errors.New("match proxy list: noting found")
	}

	var result []string
	for _, m := range proxies {
		result = append(result, fmt.Sprintf("%s://%s:%s", strings.ToLower(string(m[3])), string(m[1]), string(m[2])))
	}
	return result, pageNum, nil
}
