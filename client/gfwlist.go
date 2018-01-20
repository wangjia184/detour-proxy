package client

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var lastError string

type gfwListRule interface {
	init(rule string) error
	match(host string) bool
}

type hostUrlWildcardRule struct {
	only_http bool
	host_rule string
	url_rule  string
}

func (r *hostUrlWildcardRule) init(rule string) (err error) {
	if !strings.Contains(rule, "/") {
		r.host_rule = rule
		return
	}
	rules := strings.SplitN(rule, "/", 2)
	r.host_rule = rules[0]
	if nil != err {
		return
	}
	if len(rules) == 2 && len(rules[1]) > 0 {
		r.url_rule = rules[1]
	}
	return
}

func (r *hostUrlWildcardRule) match(host string) bool {
	if ret := wildcardMatch(host, r.host_rule); !ret {
		return false
	}
	return true
}

type urlWildcardRule struct {
	url_rule string
}

func (r *urlWildcardRule) init(rule string) (err error) {
	if !strings.Contains(rule, "*") {
		rule = "*" + rule
	}
	r.url_rule = rule
	return
}

func (r *urlWildcardRule) match(host string) bool {
	return wildcardMatch(host, r.url_rule)
}

type urlRegexRule struct {
	is_raw_regex bool
	url_reg      *regexp.Regexp
}

func (r *urlRegexRule) init(rule string) (err error) {
	if r.is_raw_regex {
		r.url_reg, err = regexp.Compile(rule)
	} else {
		r.url_reg, err = prepareRegexp(rule, false)
	}
	return
}

func (r *urlRegexRule) match(host string) bool {
	ret := r.url_reg.MatchString(host)
	//	if ret{
	//	   log.Printf("url is %s, rule is %s\n", util.GetURLString(req, false), r.url_reg.String())
	//	}
	return ret
}

type GFWList struct {
	white_list []gfwListRule
	black_list []gfwListRule
}

func (gfw *GFWList) IsBlocked(host string) bool {
	for _, rule := range gfw.white_list {
		if rule.match(host) {
			return false
		}
	}
	for _, rule := range gfw.black_list {
		if rule.match(host) {
			//log.Printf("matched for :%v for %v\n", req.URL, rule)
			return true
		}
	}
	return false
}

func Parse(rules string) (*GFWList, error) {
	reader := bufio.NewReader(strings.NewReader(rules))
	gfw := new(GFWList)
	//i := 0
	for {
		line, _, err := reader.ReadLine()
		if nil != err {
			break
		}
		str := strings.TrimSpace(string(line))
		//comment
		if strings.HasPrefix(str, "!") || len(str) == 0 {
			continue
		}
		if strings.HasPrefix(str, "@@||") {
			rule := new(hostUrlWildcardRule)
			err := rule.init(str[4:])
			if nil != err {
				log.Printf("Failed to init exclude rule:%s for %v\n", str[4:], err)
				continue
			}
			gfw.white_list = append(gfw.white_list, rule)
		} else if strings.HasPrefix(str, "||") {
			rule := new(hostUrlWildcardRule)
			err := rule.init(str[2:])
			if nil != err {
				log.Printf("Failed to init host url rule:%s for %v\n", str[2:], err)
				continue
			}
			gfw.black_list = append(gfw.black_list, rule)
		} else if strings.HasPrefix(str, "|http") {
			rule := new(urlWildcardRule)
			err := rule.init(str[1:])
			if nil != err {
				log.Printf("Failed to init url rule:%s for %v\n", str[1:], err)
				continue
			}
			gfw.black_list = append(gfw.black_list, rule)
		} else if strings.HasPrefix(str, "/") && strings.HasSuffix(str, "/") {
			rule := new(urlRegexRule)
			rule.is_raw_regex = true
			err := rule.init(str[1 : len(str)-1])
			if nil != err {
				log.Printf("Failed to init url rule:%s for %v\n", str[1:len(str)-1], err)
				continue
			}
			gfw.black_list = append(gfw.black_list, rule)
		} else {
			rule := new(hostUrlWildcardRule)
			rule.only_http = true
			err := rule.init(str)
			if nil != err {
				log.Printf("Failed to init host url rule:%s for %v\n", str, err)
				continue
			}
			gfw.black_list = append(gfw.black_list, rule)
		}
	}
	return gfw, nil
}

func ParseRawGFWList(rules string) (*GFWList, error) {
	content, err := base64.StdEncoding.DecodeString(string(rules))
	if err != nil {
		return nil, err
	}
	return Parse(string(content))
}

func wildcardMatch(text string, pattern string) bool {
	cards := strings.Split(pattern, "*")
	for _, str := range cards {
		idx := strings.Index(text, str)
		if idx == -1 {
			return false
		}
		text = strings.TrimLeft(text, str+"*")
	}
	return true
}

func getURLString(req *http.Request, with_method bool) string {
	if nil == req {
		return ""
	}
	str := req.URL.String()
	if len(req.URL.Scheme) == 0 && strings.EqualFold(req.Method, "Connect") && len(req.URL.Path) == 0 {
		str = fmt.Sprintf("https://%s", req.Host)
	}
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
		scheme := req.URL.Scheme
		if len(req.URL.Scheme) == 0 {
			scheme = "http"

		}
		str = fmt.Sprintf("%s://%s%s", scheme, req.Host, str)
	}
	if with_method {
		return fmt.Sprintf("%s %s", req.Method, str)
	}
	return str
}

func prepareRegexp(rule string, only_star bool) (*regexp.Regexp, error) {
	rule = strings.TrimSpace(rule)
	rule = strings.Replace(rule, ".", "\\.", -1)
	if !only_star {
		rule = strings.Replace(rule, "?", "\\?", -1)
	}
	rule = strings.Replace(rule, "*", ".*", -1)
	return regexp.Compile(rule)
}
