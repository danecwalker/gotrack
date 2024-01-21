package event

type Event struct {
	Session       string                 `json:"s"`
	Name          string                 `json:"n"`
	Referrer      string                 `json:"r"`
	Url           string                 `json:"u"`
	Domain        string                 `json:"d"`
	MaxNumTouches int                    `json:"t"`
	Meta          map[string]interface{} `json:"m"`
	Props         map[string]interface{} `json:"p"`
	Revenue       *Revenue               `json:"$"`
	Browser       string
	OS            string
	Device        DeviceType
	UTM           *UTM
}

type UTM struct {
	Source   string `json:"utm_source"`
	Medium   string `json:"utm_medium"`
	Campaign string `json:"utm_campaign"`
	Term     string `json:"utm_term"`
	Content  string `json:"utm_content"`
	Referrer string `json:"ref"`
}

type DeviceType string

const (
	Desktop DeviceType = "desktop"
	Mobile  DeviceType = "mobile"
	Tablet  DeviceType = "tablet"
)

type Revenue struct {
	Amount   float64 `json:"a"`
	Currency string  `json:"c"`
}
