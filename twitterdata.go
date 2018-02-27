package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	Conf "github.com/mattga/SocialData/Config"
)

func main() {
	anaconda.SetConsumerKey(<twitter key>)
	anaconda.SetConsumerSecret(<twitter secret>)
	api := anaconda.NewTwitterApi(<>,<>)
	api.EnableThrottling(15*time.Minute/180, 10000) // 180 Queries / 15 Min

	for _, hashtag := range Conf.Hashtags {
		v := url.Values{}
		v.Set("count", "100")
		v.Set("include_entities", "0")
		v.Set("lang", "en")
		v.Set("since", "2015-11-16")

		result, err := api.GetSearch(fmt.Sprintf("%%23%s -filter:retweets", hashtag), v)
		if err != nil {
			fmt.Println(err)
		}

		count := result.Metadata.Count
		fmt.Printf("New query for tweets with #%s (%d) \n", hashtag, count)

		f, err := os.OpenFile(fmt.Sprintf("data/%s.txt", hashtag), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}

		for k := 2; ; k++ {
			for _, tweet := range result.Statuses {
				text := strings.Replace(tweet.Text, "\n", " ", -1)
				data := fmt.Sprintf("%s, %s, %s\n", tweet.IdStr, tweet.CreatedAt, text)

				if _, err = f.WriteString(data); err != nil {
					panic(err)
				}
			}

			if result.Metadata.NextResults == "" {
				break
			}
			if result, err = result.GetNext(api); err != nil {
				panic(err)
			}

			count = count + result.Metadata.Count
			fmt.Printf("\tFetched tweets up to id %s (%d)\n", result.Metadata.MaxIdString, count)
		}

		f.Close()
	}
}
