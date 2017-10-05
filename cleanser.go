package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"flag"
	Conf "github.com/mattga/SocialData/Config"
)

var (
//	dataSrc                            = "twitter"
//	dataDir                            = "data/twitter11_6-15"
	dataSrc                            = "youtube"
	dataDir                            = "data/youtube11_26"
	wg                                 = sync.WaitGroup{} // This is like a semaphore
	corpus1, corpus2, corpus3, corpus4 *os.File
	count                              = 0
)

func main() {
	Conf.Load()

	var src = *flag.String("source", dataSrc, "Data source: twitter or youtube")
	var dir = *flag.String("data-dir", dataDir, "Directory containing twitter or youtube data files")
	var corp = *flag.Bool("corpus", false, "Cleanse corpus data")
	var fmtStr string

	fmt.Println(src, dir, corp)

	var terms []string
	if src == "twitter" {
		terms = Conf.Hashtags
		fmtStr = "%s"
	} else if src == "youtube" {
		terms = Conf.YTSearchTerms
		fmtStr = "yt_%s"
	}

	var err error
	corpusFile1 := fmt.Sprintf("%s/CORPUS1.txt", dataDir)
	if corpus1, err = os.OpenFile(corpusFile1, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		panic(err)
	}
	corpusFile2 := fmt.Sprintf("%s/CORPUS2.txt", dataDir)
	if corpus2, err = os.OpenFile(corpusFile2, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		panic(err)
	}
	corpusFile3 := fmt.Sprintf("%s/CORPUS3.txt", dataDir)
	if corpus3, err = os.OpenFile(corpusFile3, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		panic(err)
	}
	corpusFile4 := fmt.Sprintf("%s/CORPUS4.txt", dataDir)
	if corpus4, err = os.OpenFile(corpusFile4, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		panic(err)
	}

	// Cleanse all files in parallel
	//			for _, hashtag := range []string{"test"} {
	for _, file := range terms {
		file = fmt.Sprintf(fmtStr, strings.Replace(file, " ", "-", -1))
		fmt.Println("Cleansing", fmt.Sprintf("%s/%s.txt", dir, file))
		wg.Add(1)            // Add 1
		go cleanseData(file) // 'go' runs this function on a new thread
	}

	if corp {
		cleanseCorpus("CORPUS")
	}

	wg.Wait() // Wait for all threads
	fmt.Println("Finished cleansing", count, "lines")
}

// cleanseCorpus runs cleaning on corpus during which each line is assumed to be classified
// with <->, <+>, or <_> at the beginning for negative, positive, and neutral sentiment.
func cleanseCorpus(file string) {
	var (
		fin, fout *os.File
		err       error
		size      = 0
		posCount  = 0
		negCount  = 0
		neuCount  = 0
	)

	inFile := fmt.Sprintf("%s/%s.txt", dataDir, file)
	if fin, err = os.Open(inFile); err != nil {
		panic(err)
	}

	outFile := fmt.Sprintf("%s/cleansed/%s_cleansed.txt", dataDir, file)
	if fout, err = os.OpenFile(outFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		panic(err)
	}

	senti := make([]string, 0)
	tweets := make([]string, 0)
	reader := bufio.NewReader(fin)
	for {
		line, isPrefix, err := reader.ReadLine()
		fmt.Println(string(line))
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		} else if isPrefix {
			fmt.Println("isPrefix true...")
		}

		data := strings.SplitN(string(line), " ", 2)
		if data[0] == "" {
			break
		} else if data[0][1] == '-' {
			senti = append(senti, "negative")
			negCount += 1
		} else if data[0][1] == '+' {
			senti = append(senti, "positive")
			posCount += 1
		} else {
			senti = append(senti, "neutral")
			neuCount += 1
		}
		tweets = append(tweets, data[1])

		size = size + 1
	}
	count += size

	keywords, urlCounts, posEmoji, negEmoji := cleanseAndTokenize(tweets)

	for i := 0; i < size; i++ {
		out := fmt.Sprintf("<%s> %s <#urls:%d> <+emoji:%d> <-emoji:%d>\n",
			senti[i],
			strings.Join(keywords[i], " "),
			urlCounts[i],
			posEmoji[i],
			negEmoji[i],
		)

		fout.WriteString(out)
	}

	fmt.Printf("Done cleansing %s (%d lines; %d positive, %d negative, %d neutral)\n",
		inFile, count, posCount, negCount, neuCount)
}

// cleanseData performed reading of the data, and writing of the cleansed output with some
// additional formatting and statistics
func cleanseData(file string) {
	var (
		fin, fout *os.File
		err       error
		size      = 0
	)

	inFile := fmt.Sprintf("%s/%s.txt", dataDir, file)
	if fin, err = os.Open(inFile); err != nil {
		panic(err)
	}

	outFile := fmt.Sprintf("%s/cleansed/%s_cleansed.txt", dataDir, file)
	if fout, err = os.OpenFile(outFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600); err != nil {
		panic(err)
	}

	tweets := make([]string, 0)
	dates := make([]string, 0)
	reader := bufio.NewReader(fin)
	for {
		line, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		for isPrefix {
			var _line []byte
			_line, isPrefix, _ = reader.ReadLine()
			line = append(line, _line...)
		}

		data := strings.SplitN(string(line), ", ", 3)
		dates = append(dates, data[1])
		tweets = append(tweets, data[2])

		size += +1
	}
	count += size

	keywords, urlCounts, posEmoji, negEmoji := cleanseAndTokenize(tweets)

	freq := int(1 / .02)
	last := 1
	k := 0
	for i := 0; i < size; i++ {
		if !reflect.DeepEqual(keywords[i], keywords[last]) { // ignore consecutive duplicates
			last = i
			out := fmt.Sprintf("%s <#urls:%d> <+emoji:%d> <-emoji:%d>\n",
				strings.Join(keywords[i], " "),
				urlCounts[i],
				posEmoji[i],
				negEmoji[i],
			)

			fout.WriteString(out)

			if i%freq == 0 {
				out = fmt.Sprintf("<> %s\n", tweets[i])
				switch k % 4 {
				case 0:
					corpus1.WriteString(out)
				case 1:
					corpus2.WriteString(out)
				case 2:
					corpus3.WriteString(out)
				case 3:
					corpus4.WriteString(out)
				}
				k += 1
			}
		}
	}

	fmt.Println("Done cleansing", inFile)
	wg.Done() // Subtracts 1
}

// cleanseAndTokenize cleans an array of strings of html characters, emoticons,urls, punctuation,
// whitespace, and stopwords. Some additional cleansing we decided not to use is commented.
func cleanseAndTokenize(tweets []string) (keywords [][]string, urlCounts, posEmoji, negEmoji []int8) {
	len := len(tweets)
	keywords = make([][]string, len)
	urlCounts = make([]int8, len)
	posEmoji = make([]int8, len)
	negEmoji = make([]int8, len)

	//	for i := 0; i < 100; i++ {
	//		fmt.Print(" ")
	//	}
	//	fmt.Println("| Done")

	var regex *regexp.Regexp
	for i, tweet := range tweets {
		//		fmt.Println(tweet)
		//		if i > 0 && math.Floor(float64(i-1)*100./float64(len)) < math.Floor(float64(i)*100./float64(len)) {
		//			fmt.Print("-")
		//		}

		// Unescape HTML special characters like &amp;
		s := []byte(html.UnescapeString(tweet))

		// Remove hashtags
		//		regex, _ = regexp.Compile("#[\\S]+")
		//		s = regex.ReplaceAll(s, []byte(""))
		//		fmt.Println("HASHTAGS", string(s))

		// Remove @user
		//		regex, _ = regexp.Compile("@[\\S]+")
		//		s = regex.ReplaceAll(s, []byte(""))
		//		fmt.Println("USERS", string(s))

		// Count and remove urls
		s, urlCounts[i] = cleanseURLs(s)
		//		fmt.Println("URLS", string(s))

		// Cleanse emoticons
		s, posEmoji[i], negEmoji[i] = cleanseEmoticons(s)
		//		fmt.Println("EMOTICONS", string(s))

		// Remove lone . and ,
		//		s = cleanseSingleCommaDot(s)

		// Remove all other characters
		regex, _ = regexp.Compile("[^a-zA-Z0-9!.,?%#@ ]*")
		s = regex.ReplaceAll(s, []byte(""))
		//		fmt.Println("CHARS", string(s))

		// Whitespace -> one space
		regex, _ = regexp.Compile("\\s+")
		s = regex.ReplaceAll(s, []byte(" "))
		//		fmt.Println("SPACE", string(s))

		keywords[i] = removeStopwords(s)
		//		fmt.Println("DONE", keywords[i])
	}

	return
}

// cleanseURLs counts and removes URLs from a string
func cleanseURLs(s []byte) (s2 []byte, urlCount int8) {
	s2 = s
	//	regex, _ := regexp.Compile("http(s?)://t.co/[a-zA-Z0-9]+")
	//	regex, _ := regexp.Compile("http(s?)://[a-zA-Z0-9.-]+/[a-zA-Z0-9.=%?/#]*")
	regex, _ := regexp.Compile("((http|https):\\/\\/?)[a-zA-Z0-9-]+\\.[a-zA-Z0-9.=%?/#]*")
	for {
		loc := regex.FindIndex(s2)
		if loc == nil {
			break
		}
		urlCount += 1

		s2 = append(s2[:loc[0]], s2[loc[1]:]...)
	}

	return
}

// cleanseEmoticons formats emoticons (if any) into unicode, then counts emoticons based on
// their sentiment which we query for from our own dictionary (loaded from data/sentiment/emoticons.txt)
func cleanseEmoticons(s []byte) (s2 []byte, pos, neg int8) {
	s2 = []byte(fmt.Sprintf("%+q", s)) // Puts emoticons in utf-16 format
	s2 = s2[1 : len(s2)-1]             // Remove quotes
	//	fmt.Println(string(s2))

	regex, _ := regexp.Compile("\\\\U0*[0-9a-z]+|\\\\u0*[0-9a-z]+")
	loc := regex.FindIndex(s2)
	for loc != nil {
		emoji := s2[loc[0]:loc[1]]

		regex2, _ := regexp.Compile("\\\\U0*|\\\\u0*")
		hexcode := string(regex2.ReplaceAll(emoji, []byte("")))
		senti := Conf.EmoticonSenti[strings.ToUpper(hexcode)]
		if senti != 0 {
			if senti > 0 {
				pos += 1
			} else {
				neg += 1
			}
		}

		s2 = append(s2[:loc[0]], s2[loc[1]:]...)
		loc = regex.FindIndex(s2)
	}

	return
}

// cleanseSingleCommaDot removes all single instances of commas and dots, leaving instances with 1 or more
// consecutive dots/commas
func cleanseSingleCommaDot(s []byte) (s2 []byte) {
	s2 = s
	len := len(s2)
	for i, b := range s2 {
		if b == 46 || b == 44 {
			if i-1 > 0 && (s2[i-1] != 46 && s2[i-1] != 44) {
				if i+1 < len && (s2[i+1] != 46 && s2[i+1] != 44) {
					s2[i] = 32
				}
			}
		}
	}

	return
}

// removeStopwords simple removes all words that appear in the StopWords variable
// of Config/Config.go
func removeStopwords(s []byte) (res []string) {
	res = make([]string, 0)
	keywords := strings.Split(string(s), " ")

	if len(keywords) <= 1 {
		return
	}

	if keywords[len(keywords)-1] == "" {
		keywords = keywords[:len(keywords)-1]
	}

	if keywords[0] == "" {
		keywords = keywords[1:]
	}

	for _, word := range keywords {
		if Conf.StopWords[word] == 0 {
			res = append(res, word)
		}
	}
	res = keywords

	return
}
