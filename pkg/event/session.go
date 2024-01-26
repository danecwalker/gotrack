package event

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mileusna/useragent"
)

type DeviceType string

const (
	Mobile  DeviceType = "mobile"
	Tablet  DeviceType = "tablet"
	Laptop  DeviceType = "laptop"
	Desktop DeviceType = "desktop"
	Display DeviceType = "display"
)

var BreakPoints = []int{640, 768, 1024, 1280, 1536}

type Session struct {
	Language   string
	Country    string
	Browser    string
	Os         string
	DeviceType DeviceType
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) ParseUA(ua string, platform string, browser string) {
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
		s.DeviceType = Desktop
		return
	}

	w := 0
	h := 0
	_, err := fmt.Sscanf(size, "%dx%d", &w, &h)
	if err != nil {
		s.DeviceType = Desktop
		return
	}

	if w < BreakPoints[0] {
		s.DeviceType = Mobile
	} else if w < BreakPoints[1] {
		s.DeviceType = Tablet
	} else if w < BreakPoints[2] {
		s.DeviceType = Laptop
	} else if w < BreakPoints[3] {
		s.DeviceType = Desktop
	} else {
		s.DeviceType = Display
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
