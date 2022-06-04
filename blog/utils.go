package blog

import (
	"encoding/base64"
	"fmt"
	"strings"
)

/*
 * Split the markdown with <!--more--> as the summary
 */
func SplitSummary(md string) (string, error) {
	md = strings.Replace(md, "&lt;!--more--&gt;", "<!--more-->", -1)
	contents := strings.Split(md, "<!--more-->")
	if len(contents) == 1 {
		return "No Summary", fmt.Errorf("no <!--more--> tag")
	} else {
		summary := contents[0]
		return summary, nil
	}
}

func Inc(i int) int {
	return i + 1
}

func PrepareContent(content string) string {
	_content, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "Load content error"
	} else {
		_text_content := strings.Replace(string(_content), "&lt;!--more--&gt;", "<!--more-->", -1)
		return _text_content
	}
}
