//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"fmt"
	"time"
)

func producerAsync(stream Stream) <-chan *Tweet {
	tweets := make(chan *Tweet)

	go func() {
		defer close(tweets)

		for {
			tweet, err := stream.Next()

			if err == ErrEOF {
				return
			}

			tweets <- tweet

		}

	}()

	return tweets
}

func consumerAsync(tweets <-chan *Tweet) {

	for tweet := range tweets {
		if tweet.IsTalkingAboutGo() {
			fmt.Println(tweet.Username, "\ttweets about golang")
			continue
		}

		fmt.Println(tweet.Username, "\tdoes not tweet about golang")
	}
}

func producer(stream Stream) (tweets []*Tweet) {
	for {
		tweet, err := stream.Next()
		if err == ErrEOF {
			return tweets
		}

		tweets = append(tweets, tweet)
	}
}

func consumer(tweets []*Tweet) {
	for _, t := range tweets {
		if t.IsTalkingAboutGo() {
			fmt.Println(t.Username, "\ttweets about golang")
		} else {
			fmt.Println(t.Username, "\tdoes not tweet about golang")
		}
	}
}

func syncProducerConsume() {
	start := time.Now()
	stream := GetMockStream()
	tweets := producer(stream)
	consumer(tweets)
	fmt.Printf("Sync Process took %s\n", time.Since(start))
}

func asyncProducerConsume() {
	start := time.Now()
	stream := GetMockStream()

	tweets := producerAsync(stream)
	consumerAsync(tweets)

	fmt.Printf("Async Process took %s\n", time.Since(start))

}

func main() {
	asyncProducerConsume()
}
