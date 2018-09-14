package proxy

// http://www.mayidaili.com/free/2
import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"time"
)

var ip66re *regexp.Regexp

func init() {
	ip66re = regexp.MustCompile(`\d+\.\d+\.\d+\.\d+:\d+`)
}

type ip66 struct{}

func (s *ip66) Wait() {
	time.Sleep(5 * time.Second)
}

func (s *ip66) Fetch(page int) ([]string, int, error) {
	client := http.DefaultClient
	resp, err := client.Get(fmt.Sprintf("http://www.66ip.cn/mo.php?tqsl=100"))
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

	result := ip66re.FindAll(data, -1)
	var list []string
	for _, line := range result {
		list = append(list, fmt.Sprintf("http://%s", string(line)))
	}
	return list, math.MaxInt64, nil
}
