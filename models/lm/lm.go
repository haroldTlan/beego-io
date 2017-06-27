package lm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"speedio/util"
)

type LocationMapping struct {
	Mapping map[string]string
	DsuList map[string]int
}

var (
	company = beego.AppConfig.String("company")
	sn, _   = beego.AppConfig.Int("slotNum")
)

func init() {

}

func (l *LocationMapping) Init() {
	l.Mapping = make(map[string]string, 0)
	l.DsuList = make(map[string]int, 0)

	cmd := "ls /sys/class/scsi_host/ | wc -l"
	o, err := util.ExecuteByStr(cmd, true)
	if err != nil {
		return
	}

	o = strings.TrimSpace(o)

	if c, _ := strconv.Atoi(o); c > 10 {
		nodisk := systemDisk()
		cmd = fmt.Sprintf("ls -l /sys/block/ |grep 'ata' |grep -v '%s'", nodisk[0:3])
		o, err := util.ExecuteByStr(cmd, true)
		if err != nil {
			return
		}
		o = strings.TrimSpace(o)

		for _, n := range strings.Split(o, "\n") {
			m := regexp.MustCompile(`ata(\d+)`)
			if m.MatchString(n) {
				// group 1
				mList := m.FindSubmatch([]byte(n))
				portid := string(mList[1])
				mm := regexp.MustCompile(`block/(\w+)`)

				if mm.MatchString(n) {
					mmList := mm.FindSubmatch([]byte(n))
					disk := string(mmList[1])
					if status, err := checkDisk(disk); status && err == nil {
						l.Mapping[fmt.Sprintf("1.1.%s", RealToLogicMapping(company, sn)[portid])] = disk
					}
				}
			}
		}
		l.DsuList["1.1"] = sn
		//	fmt.Printf("???%+v", l.Mapping)
		return
	} else if c > 8 {
		nodisk := systemDisk()
		cmd = fmt.Sprintf("ls -l /sys/block/ |grep 'ata' |grep -v '%s'", nodisk[0:3])
		o, err := util.ExecuteByStr(cmd, true)
		if err != nil {
			return
		}
		o = strings.TrimSpace(o)

		for _, n := range strings.Split(o, "\n") {
			m := regexp.MustCompile(`target(\d+):(\d+)`)
			if m.MatchString(n) {
				// group 2
				mList := m.FindSubmatch([]byte(n))
				portid1 := string(mList[1])
				portid2 := string(mList[2])
				mm := regexp.MustCompile(`block/(\w+)`)

				if mm.MatchString(n) {
					portid := portid1 + "-" + portid2
					mmList := mm.FindSubmatch([]byte(n))
					disk := string(mmList[1])
					if status, err := checkDisk(disk); status && err == nil {
						l.Mapping[fmt.Sprintf("1.1.%s", RealToLogicMapping(company, sn)[portid])] = disk
					}
				}
			}
		}
		l.DsuList["1.1"] = 16
		//	fmt.Printf("%+v", l.Mapping)
		return
	}

	for loc, dev := range RealToLogicMapping("normal", 16) {
		if _, err := os.Stat("/dev/" + dev); os.IsNotExist(err) {
			continue
		}
		l.Mapping[loc] = dev
	}
	l.DsuList["1.1"] = 16

	return
	//fmt.Printf("%+v\n", l.Mapping)
	/*
		cmd = "ls -l /sys/block/ |grep host |grep -v sda"
		o, err = util.ExecuteByStr(cmd, true)
		if err != nil || o == "" {
			return
		}

		num, chanel := _checkExpander()
		if num > 0 {
			for _, ch := range chanel {
				fmt.Println(ch)
				//		makeMapping()
			}
		} else {

			cmd = "ls -l /sys/block/ |grep host |grep -v sda"
			o, err := util.ExecuteByStr(cmd, true)
			if err != nil {
				return
			}
			o = strings.TrimSpace(o)

			/*		for _, nr := range strings.Split(o, "\n") {
					n := strings.Split(nr, "")
					if len(n) == 11 {
						blk := n[8]
						loc := n[10]
						m := regexp.MustCompile(`port-\d+:\d+`)

				}
		}*/

}

func makeMapping() {
	/*slot := 24
	num := 0
	diskList := make([]string, 0)
	dsu := "1.1"

	cmd := "ls -l /sys/block/ |grep host |grep -v sda"
	o, err := util.ExecuteByStr(cmd, true)
	if err != nil {
		return
	}
	o = strings.TrimSpace(o)

	for _, n := range strings.Split(o, "\n") {
		//	c:="expander"
	}*/
}

func _checkExpander() (chanelCount int, chanelList []string) {
	chanelMap := make(map[string]bool, 0)
	cmd := "ls /sys/class/sas_expander"
	o, err := util.ExecuteByStr(cmd, true)
	if err != nil {
		return
	}
	o = strings.TrimSpace(o)
	if o == "" {
		return
	}
	for _, n := range strings.Split(o, "\n") {
		m := regexp.MustCompile(`expander-\d+:(\d+)`)
		if m.MatchString(n) {
			mList := m.FindSubmatch([]byte(n))
			chanel := string(mList[1])
			if ok := chanelMap[chanel]; !ok {
				chanelMap[chanel] = true
				chanelList = append(chanelList, chanel)
			}

		}
	}
	return
}

func systemDisk() (sd string) {
	nd, err := filepath.Glob("/dev/sd?3")
	if err != nil {
		util.AddLog(err)
		return
	}

	for _, d := range nd {
		cmd := fmt.Sprintf("blockdev --getsz /dev/%s", d[5:8])
		o, err := util.ExecuteByStr(cmd, true)
		if err != nil {
			return
		}
		o = strings.TrimSpace(o)
		if c, _ := strconv.Atoi(o); c < 72533296 {
			return d[5:8]
		}
	}

	return "sdabc"
}

func checkDisk(disk string) (bool, error) {
	cmd := fmt.Sprintf("blockdev --getsz /dev/%s", disk)
	o, err := util.ExecuteByStr(cmd, false)
	if err != nil {
		err := fmt.Errorf("no such file")
		util.AddLog(err) //TODO
		return false, err
	}
	o = strings.TrimSpace(o)

	if c, _ := strconv.Atoi(o); c < 72533296*4 {
		return false, nil
	}
	return true, nil
}

func RealToLogicMapping(company string, slotNumbers int) map[string]string {
	if company == "qinchen" && slotNumbers == 24 {
		RealToLogicMap := map[string]string{
			"1":  "21",
			"2":  "22",
			"3":  "23",
			"4":  "24",
			"5":  "0",
			"6":  "0",
			"7":  "1",
			"8":  "2",
			"9":  "3",
			"10": "4",
			"11": "5",
			"12": "6",
			"13": "7",
			"14": "8",
			"15": "9",
			"16": "10",
			"17": "11",
			"18": "12",
			"19": "13",
			"20": "14",
			"21": "15",
			"22": "16",
			"23": "17",
			"24": "18",
			"25": "19",
			"26": "20",
		}
		return RealToLogicMap
	} else if company == "gooxi" && slotNumbers == 24 {
		RealToLogicMap := map[string]string{
			"0":  "0",
			"1":  "1",
			"2":  "2",
			"3":  "3",
			"4":  "4",
			"5":  "5",
			"6":  "6",
			"7":  "7",
			"8":  "8",
			"9":  "9",
			"10": "10",
			"11": "11",
			"12": "12",
			"13": "13",
			"14": "14",
			"15": "15",
			"16": "16",
			"17": "17",
			"18": "18",
			"19": "19",
			"20": "20",
			"21": "21",
			"22": "22",
			"23": "23",
		}
		return RealToLogicMap
	} else if company == "qinchen" && slotNumbers == 16 {
		RealToLogicMap := map[string]string{
			"12": "0",
			"13": "1",
			"14": "2",
			"15": "3",
			"8":  "4",
			"9":  "5",
			"10": "6",
			"11": "7",
			"4":  "8",
			"5":  "9",
			"6":  "10",
			"7":  "11",
			"0":  "12",
			"1":  "13",
			"2":  "14",
			"3":  "15",
		}
		return RealToLogicMap
	} else if company == "normal" && slotNumbers == 16 {
		RealToLogicMap := map[string]string{
			"1.1.2":  "sdc",
			"1.1.3":  "sdd",
			"1.1.4":  "sde",
			"1.1.5":  "sdf",
			"1.1.6":  "sdg",
			"1.1.7":  "sdh",
			"1.1.8":  "sdi",
			"1.1.9":  "sdj",
			"1.1.10": "sdk",
			"1.1.11": "sdl",
			"1.1.12": "sdm",
			"1.1.13": "sdn",
			"1.1.14": "sdo",
			"1.1.15": "sdp",
			"1.1.16": "sdq",
		}
		return RealToLogicMap
	} else {
		RealToLogicMap := map[string]string{
			"0":  "0",
			"1":  "1",
			"2":  "2",
			"3":  "3",
			"4":  "4",
			"5":  "5",
			"6":  "6",
			"7":  "7",
			"8":  "8",
			"9":  "9",
			"10": "10",
			"11": "11",
			"12": "12",
			"13": "13",
			"14": "14",
			"15": "15",
		}
		return RealToLogicMap
	}
}
