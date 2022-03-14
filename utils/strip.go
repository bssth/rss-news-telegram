package html

import "regexp"

const regex = `<.*?>`

func StripHTML(s string) string {
	r := regexp.MustCompile(regex)
	return r.ReplaceAllString(s, "")
}
