package lm

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
			"1.1.1":  "sdb",
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
