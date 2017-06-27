package util

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
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

func Format(format, uuid string) string {
	m := regexp.MustCompile(`\d+(.)`)
	sep := m.FindSubmatch([]byte(format))[1]

	var segmentNr []int
	for _, nr := range strings.Split(format, string(sep)) {
		num, _ := strconv.Atoi(nr)
		segmentNr = append(segmentNr, num)
	}

	mm := regexp.MustCompile(`\w+(.?)`)
	var sep1 string
	if str := mm.FindSubmatch([]byte(uuid)); len(str) > 0 {
		sep1 = string(str[1])
	} else {
		sep1 = ""
	}

	uuid = strings.Replace(uuid, sep1, "", -1)
	var segment []string
	for _, nr := range segmentNr {
		var seg string
		seg, uuid = uuid[0:nr], uuid[nr:]
		segment = append(segment, seg)
	}

	return strings.Join(segment, string(sep))
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
