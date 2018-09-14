package command

import "regexp"

var google = "https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=zh-CN&hl=zh-CN&dt=t&dt=bd&dj=1&source=input&tk=54722.54722"

var entitiesCounter, word, exporter, foreign, ref, originImg *regexp.Regexp

func init() {
	entitiesCounter = regexp.MustCompile(`<div\sclass="searchList__.*">(\d+)\sentries found</div>`)
	word = regexp.MustCompile(`<p\sclass="word__name.*?"\stitle=".*?">(.*?)\s?</p>`)
	exporter = regexp.MustCompile(`<div class="word--.*?">.*?title=".*?">(.*?)<.*?object>(?s:(.*?))</object>`)
	foreign = regexp.MustCompile(`foreign">(.*?)</span>`)
	ref = regexp.MustCompile(`<a\shref=.*?crossreference.*?>(.*?)</a>`)
	originImg = regexp.MustCompile(`<img\s+id="lr_dct_img_origin.*?"\s+src="(data.*?)"`)
}
