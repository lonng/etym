package proxy

// http://www.mayidaili.com/free/2
import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var u5re *regexp.Regexp

func init() {
	u5re = regexp.MustCompile(`<li>(\d+\.\d+\.\d+\.\d+)</li>(?s:.*?)<li\sclass="port.*?">(\d+)</li>(?s:.*?)>(https?)</a>`)
}

type u5 struct{}

func (s *u5) Wait() {
	time.Sleep(60 * time.Second)
}

func (s *u5) Fetch(page int) ([]string, int, error) {
	req, err := http.NewRequest(http.MethodGet, "http://www.data5u.com/free/index.shtml", nil)
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

	proxies := u5re.FindAllSubmatch(data, -1)
	if len(proxies) < 1 {
		return nil, 0, errors.New("match proxy list: noting found")
	}

	var result []string
	for _, m := range proxies {
		result = append(result, fmt.Sprintf("%s://%s:%s", strings.ToLower(string(m[3])), string(m[1]), string(m[2])))
	}
	return result, math.MaxInt64, nil
}
