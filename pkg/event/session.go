package event

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mileusna/useragent"
)

type ScreenType string

const (
	Mobile  ScreenType = "mobile"
	Tablet  ScreenType = "tablet"
	Laptop  ScreenType = "laptop"
	Desktop ScreenType = "desktop"
	Display ScreenType = "display"
)

var BreakPoints = []int{640, 768, 1024, 1280, 1536}

type Session struct {
	SessionID  string
	Language   string
	Country    string
	Browser    string
	Os         string
	ScreenType ScreenType
}

func NewSession(r *http.Request) *Session {
	ip := getIP(r)
	sha := sha1.New()
	sha.Write([]byte(ip))
	sha.Write([]byte(r.Header.Get("User-Agent")))
	sha.Write([]byte(fmt.Sprintf("%d", time.Now().Unix()/60/60/24)))
	return &Session{
		SessionID: fmt.Sprintf("%x", sha.Sum(nil)),
	}
}

func (s *Session) ParseUA(ua string, platform string, browser string) {
	fmt.Println(ua)
	if ua != "" {
		agent := useragent.Parse(ua)
		if platform != "" {
			s.Os = strings.Trim(platform, "\"")
		} else {
			s.Os = agent.OS
		}

		s.Browser = agent.Name

		browsers := strings.Split(browser, ", ")
		re := regexp.MustCompile(`^Not.A Brand$`)
		for _, b := range browsers {
			if strings.Contains(b, ";") {
				agent := strings.Split(b, ";")
				agentName := strings.Trim(agent[0], "\"")
				if agentName != "" && !re.MatchString(agentName) {
					s.Browser = agentName
				}
			}
		}
	}
}

func (s *Session) ParseViewportSize(size string) {
	if size == "" {
		s.ScreenType = Desktop
		return
	}

	w := 0
	h := 0
	_, err := fmt.Sscanf(size, "%dx%d", &w, &h)
	if err != nil {
		s.ScreenType = Desktop
		return
	}

	if w < BreakPoints[0] {
		s.ScreenType = Mobile
	} else if w < BreakPoints[1] {
		s.ScreenType = Tablet
	} else if w < BreakPoints[2] {
		s.ScreenType = Laptop
	} else if w < BreakPoints[3] {
		s.ScreenType = Desktop
	} else {
		s.ScreenType = Display
	}
}

func (s *Session) ParseLanguage(al string) {
	if al != "" {
		lang, country := parseAcceptLanguage(al)
		if lang != "" {
			s.Language = lang
		}
		if country != "" {
			s.Country = country
		}
	}
}

func parseAcceptLanguage(al string) (string, string) {
	for _, l := range strings.Split(al, ",") {
		if strings.Contains(l, ";") {
			lang, country := parseLanguageCountry(l)
			if lang != "" && country != "" {
				return lang, country
			}
		} else {
			lang, country := parseLanguage(l)
			if lang != "" && country != "" {
				return lang, country
			}
		}
	}
	return "", ""
}

func parseLanguageCountry(l string) (string, string) {
	parts := strings.Split(l, ";")
	if len(parts) != 2 {
		return "", ""
	}
	lang := strings.TrimSpace(parts[0])
	country := strings.TrimSpace(parts[1])
	return lang, country
}

func parseLanguage(l string) (string, string) {
	parts := strings.Split(l, "-")
	if len(parts) != 2 {
		return "", ""
	}
	lang := strings.TrimSpace(parts[0])
	country := strings.TrimSpace(parts[1])
	return lang, country
}

func getIP(r *http.Request) string {
	// Get IP from X-REAL-IP or X-FORWARDED-FOR headers if present and not localhost or fallback to RemoteAddr otherwise
	ip := r.Header.Get("X-REAL-IP")
	if ip == "" {
		ip = r.Header.Get("X-FORWARDED-FOR")

		if ip == "" {
			ip = r.RemoteAddr
		}
	}

	// check if localhost or [::1] and replace with 127.0.0.1
	re := regexp.MustCompile(`^(localhost|\[::1\])(.+)$`)
	if re.MatchString(ip) {
		ip = re.ReplaceAllString(ip, "127.0.0.1")
	}
	re = regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(:.+)$`)
	if re.MatchString(ip) {
		ip = re.ReplaceAllString(ip, "$1")
	}
	fmt.Println(ip)

	return ip
}
