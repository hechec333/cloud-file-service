package discovery

import "regexp"

var ServicePrefix string = "/sd/"
var ipRegxp *regexp.Regexp = nil

func init() {
	i, err := regexp.Compile(`[0-9.]+`)
	if err != nil {
		panic(err)
	}
	ipRegxp = i
}
func Ipvalidator(src []byte) (string, bool) {
	if ipRegxp.MatchString(string(src)) == false {
		return "", false
	} else {
		return string(src), true
	}
}
