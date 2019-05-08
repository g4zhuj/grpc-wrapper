package falconmetrics

import (
	"os"
	"os/exec"
	"strings"
)

//FalconConfig config of open-falcon
type FalconConfig struct {
	Debug    bool `json:""`
	EndPoint string `json:"endpoint"`
	HostName string `json:"hostname"` //hostname of agent default is http://127.0.0.1:1988/v1/push
	Step     int64  `json:"interval"` // interval to report metrics (s)
	BaseTags string `json:"basetags"` // base tags
}

//DefaultFalconConfig default config
var DefaultFalconConfig = FalconConfig{
	HostName: "http://127.0.0.1:1988/v1/push",
	Step:     60,
	EndPoint: defaultHostname(),
}

func defaultProjectName() string {
	s, _ := exec.LookPath(os.Args[0])
	psName := ""
	if strings.Contains(s, "/") {
		ss := strings.Split(s, "/")
		psName = ss[len(ss)-1]
	} else {
		psName = s
	}
	return psName
}

func defaultHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}
