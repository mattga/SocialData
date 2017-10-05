This code is used to query YouTube and Twitter APIs for data as well as cleanse
the data. It is written in Golang because it was easiest to interact with the Web APIs.

TA: If you want to run this code, I probably would need to demo it for you (contact me at mattga@iastate.edu).
The code is simply fetching a bunch of data and doing some minor processing in cleanser.go, so frankly
I am hoping you do not bother with needing to see this this trivial code run.



The structure is as follows:

YouTube Data
------------
youtubedata.go - Establishes an OAuth client by directing you to the browser to authorize
                 access to youtube with some google account. Queries with keywords and channel ids
                 from Config/Config.go

client_secrets.json - Stores configuration variables for this application to authenticate with YouTube API.

request.token - Stores the OAuth information (tokens + expirey) received after you authorize with
                a Google account


Twitter Data
------------
twitterdata.go - Uses a library to query the Twitter API for tweets. Throttling is enabled since
                 we are limited to 180 queries / 15 min. Queries with hashtags from Config/Config.go


Cleansing
---------
cleanser.go - Performs all file reading & writing, text processing, etc. Files are expected in the
              data directory you pass in, while output will be written to a cleansed folder within that
              directory (Must be created)
              Flags:
              -source [youtube|twitter]     Specify whether to cleanse youtube or twitter data
              -data-dir <file path>         File path to the location of data files written by youtubedata.go or twitterdata.go
              -corpus                       Tells the cleanser to cleanse the corpus

Config/Config.go - Includes variables from cleansing including list of stopwords and emoticon sentiment