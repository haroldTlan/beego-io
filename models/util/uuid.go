package util

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
)

var (
	LVM_FORMAT     = "6-4-4-4-4-4-6"
	MDADM_FORMAT   = "8:8:8:8"
	SPEEDIO_FORMAT = "8-4-4-4-12"
	VALID_CHAR     = "[0-9|a-f|A-F]"
	hostUuid       = ""
)

func Validate(uuid string) bool {
	m := regexp.MustCompile(`\d+(.)`)
	sep := m.FindSubmatch([]byte(VALID_CHAR))[1]
	var char []string
	for _, nr := range strings.Split(SPEEDIO_FORMAT, string(sep)) {
		c := fmt.Sprintf("%s{%s}", VALID_CHAR, nr)
		char = append(char, c)
	}
	pattern := strings.Join(char, string(sep))
	result, _ := regexp.MatchString(pattern, uuid)
	return result
}

func HostUuid() string {
	if len(hostUuid) == 0 {
		hostPath := beego.AppConfig.String("uuid_host_path")
		if _, err := os.Stat(hostPath); err == nil {
			hostUuid = ReadFile(hostPath)
			if !Validate(hostUuid) {
				os.Remove(hostPath)
				return HostUuid()
			}
		} else {
			hostUuid = Urandom()
			WriteFile(hostPath, hostUuid)
		}
	}

	return hostUuid
}

func Uuid4() string {
	return Urandom()
}
