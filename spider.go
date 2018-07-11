package main

import (
	"net/http"
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"
	"fmt"
)

var (
	cnBlogSpider CnBlogSpider
)

type spider interface {
	UrlMatch(url string) bool
	ParsePage(html string) string
}

type CnBlogSpider struct{}

func (CnBlogSpider) ParsePage(html string) string {

	// @see https://studygolang.com/articles/8865
	reg := regexp.MustCompile(`(<div id="topics">(?s:.*))</div><a name="!comments">`)

	matches := reg.FindAllString(html, -1)

	match := matches[0]
	match = strings.Replace(match, "</div><a name=\"!comments\">", "", -1)

	return match
}

func (CnBlogSpider) UrlMatch(url string) bool {
	return strings.Contains(url, "cnblogs.com")
}

func main() {

	url := "https://www.cnblogs.com/qcloud1001/p/9293424.html"

	request, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))

	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	response, err := client.Do(request)

	if response.StatusCode != 200 || err != nil {
		panic(response.StatusCode)
	}

	spiderInstance := spiderFactory(url)

	body, _ := ioutil.ReadAll(response.Body)

	bodyString := string(body)

	html := spiderInstance.ParsePage(bodyString)

	html = removeLineBreak(html)

	html = filterHtml(html)

	fmt.Println(html)
}

func spiderFactory(url string) spider {

	if cnBlogSpider.UrlMatch(url) {
		return cnBlogSpider
	}

	panic("have no this match url symbol")
}

func filterHtml(html string) string {

	reg := regexp.MustCompile(`<[^>]*>[^<]*|</[^>]*>`)

	matches := reg.FindAllStringSubmatch(html, -1)

	removeEmptyTag(matches)

	allTagString := ""

	for _, tag := range matches {

		tagString := tag[0]

		if isScriptTag(tagString) {
			continue
		}

		if tagString != "" {
			tagString = strings.TrimSpace(tagString)

			tagString = removeTagHtmlComment(tagString)

			allTagString += tagString
		}
	}

	return allTagString
}

func isScriptTag(tag string) bool {

	if strings.Contains(tag, "<script") {
		return true
	}
	if strings.Contains(tag, "</script") {
		return true
	}

	return false
}

func removeLineBreak(html string) string {

	html = strings.Replace(html, "\n", "", -1)
	html = strings.Replace(html, "\r\n", "", -1)
	html = strings.Replace(html, "\r", "", -1)

	return html
}

func removeTagHtmlComment(tag string) string {

	if !strings.Contains(tag, "<!--") {
		return tag
	}

	reg := regexp.MustCompile(`<!--.*-->`)
	return reg.ReplaceAllString(tag, "")
}

func removeEmptyTag(tags [][]string) {

	isEndTag := func(tag string) bool {
		return strings.Contains(tag, "</")
	}

	getEndTagName := func(endTag string) string {

		reg := regexp.MustCompile(`</[^> ]*`)
		tagName := reg.FindString(endTag)

		return strings.Replace(tagName, "</", "", -1)
	}

	// just end tag will go into this func
	removeFunc := func(index int) {

		// loop up for before index
		currentTagName := getEndTagName(tags[index][0])

		for start := index - 1; start > 0; start-- {

			lastTag := tags[start][0]

			if lastTag == "" {
				continue
			}

			pattern := fmt.Sprintf("<%s[^>]*>$", currentTagName)

			match, err := regexp.MatchString(pattern, lastTag)

			if !match || err != nil {
				break
			}

			//fmt.Println("start:" + tags[start][0] + "end:"+tags[index][0])

			tags[start][0] = ""
			tags[index][0] = ""
		}
	}

	for index, value := range tags {

		tag := value[0]

		if !isEndTag(tag) {
			continue
		}

		removeFunc(index)
	}
}
