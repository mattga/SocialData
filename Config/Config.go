package Config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	Hashtags = []string{
		// Bernie Sanders
		"sanders2016",
		"bernie2016",
		"berniesanders",
		"feelthebern",
		// Hillary Clinton
		"clinton2016",
		"hillary2016",
		"hillaryclinton",
		"hillary",
		// Martin O'Malley
		"omalley2016",
		// Donald Trump
		"trump2016",
		"trump",
		"donaldtrump",
		"makeamericagreatagain",
		// Ben Carson
		"carson2016",
		"bencarson",
		// Jeb Bush
		"bush2016",
		"jebbush",
		// Ted Cruz
		"cruz2016",
		"tedcruz",
		// Marco Rubio
		"rubio2016",
		"marcorubio",
	}

	YTChannels = map[string]string{
		"USATODAY":        "UCP6HGa63sBC7-KHtkme-p-g",
		"ABCNews":         "UCBi2mrWuNuyYy4gbM6fU18Q",
		"CBS":             "UClzCn8DxRSCuMFv_WfzkcrQ",
		"AssociatedPress": "UC52X5wxOL_s5yw0dQk7NtgA",
		"FoxNewsChannel":  "UCXIJgqnII2ZOINSWNOGFThA",
		"TheNewYorkTimes": "UCqnbDFdCpuN8CMEg0VuEBqA",
		"CNN":             "UCupvZG-5ko_eiXAupbDfxWw",
	}

	YTSearchTerms = []string{
		"bernie",
		"hillary",
		"omalley",
		"trump",
		"carson",
		"ted cruz",
		"jeb bush",
		"rubio",
	}

	// Make use of hash map for constant access
	StopWords = map[string]int{
		"a": 1, "an": 1, "and": 1, "are": 1, "as": 1, "at": 1, "be": 1, "by": 1, "for": 1, "from": 1, "has": 1,
		"he": 1, "in": 1, "is": 1, "it": 1, "its": 1, "of": 1, "on": 1, "that": 1, "the": 1, "to": 1, "was": 1,
		"were": 1, "will": 1, "with": 1,
	}

	emoticonFile  = "data/sentiment/emoticons.txt"
	EmoticonSenti = make(map[string]int8, 0)
)

func loadEmoticons() {
	fin, err := os.Open(emoticonFile)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(fin)
	for {
		line, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		} else if isPrefix {
			panic(fmt.Errorf("isPrefix true..."))
		}

		data := strings.Split(string(line), ",")
		senti, _ := strconv.Atoi(data[1])
		EmoticonSenti[strings.ToUpper(data[0])] = int8(senti)
	}
}

func Load() {
	loadEmoticons()
}
