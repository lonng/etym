package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"etym/pkg/log"
	"strings"
)

const Timeout = 20

type (
	Site interface {
		Fetch(page int) (proxyList []string, totalCount int, err error)
		Wait()
	}

	checkResult struct {
		addr   string
		Reason string
	}
)

var availableSource = []Site{
	&kuaidaili{baseUrl: "https://www.kuaidaili.com/free/intr/"},
	&ip66{},
	&xicidaili{baseUrl: "http://www.xicidaili.com/nn"},
	&xicidaili{baseUrl: "http://www.xicidaili.com/nt"},
	&u5{},
	//&kuaidaili{baseUrl: "https://www.kuaidaili.com/free/inha/"},
}

func checkAlive(s string) checkResult {
	proxyUrl, err := url.Parse(s)
	if err != nil {
		return checkResult{addr: s, Reason: err.Error()}
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: Timeout * time.Second,
	}
	resp, err := client.Get("https://www.baidu.com/duty/")
	if err != nil {
		return checkResult{addr: s, Reason: err.Error()}
	}
	if resp.StatusCode != http.StatusOK {
		return checkResult{addr: s, Reason: err.Error()}
	}
	return checkResult{addr: s}
}

func filter(proxies []string, output chan<- string) {
	var wg sync.WaitGroup
	var ch = make(chan checkResult, len(proxies))
	defer close(ch)

	// check alive
	for _, s := range proxies {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			ch <- checkAlive(addr)
		}(s)
	}

	wg.Wait()
	for range proxies {
		result := <-ch
		if result.Reason != "" {
			//log.Warnf("Dead address: %v, reason=%v", result.addr, result.Reason)
			continue
		}

		output <- result.addr
		log.Debugf("Alive address success=%v", result.addr)
	}
}

func Get(output chan<- string) error {
	var wg sync.WaitGroup
	var errs = make(chan error, len(availableSource))
	defer close(errs)

	for _, source := range availableSource {
		wg.Add(1)
		go func(s Site) {
			defer wg.Done()

			log.Infof("Start fetching goroutine, source=%s", s)

			// 目前只支持快代理
			var page = 1
			for {
				proxies, totalPage, err := s.Fetch(page)
				if err != nil {
					errs <- err
					break
				}

				filter(proxies, output)

				page++
				if page > totalPage {
					page = 1
				}
				s.Wait()
			}
		}(source)
	}

	wg.Wait()

	var collectErr []string
	var errCount = len(errs)
	for i := 0; i < errCount; i++ {
		err := <-errs
		collectErr = append(collectErr, err.Error())
	}

	if len(collectErr) > 0 {
		return fmt.Errorf("combine errors: %s", strings.Join(collectErr, "\n"))
	}

	return nil
}
