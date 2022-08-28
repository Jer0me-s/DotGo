package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Clinet/discordgo-embed"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/nstratos/go-myanimelist/mal"
	"gopkg.in/headzoo/surf.v1"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}
func main() {
	Token = "" //TODO Remove this

	dg, err := discordgo.New("Bot " + Token)
	defer func(dg *discordgo.Session) {
		err := dg.Close()
		if err != nil {

			panic(err)
		}
	}(dg) //Close the session when the main function ends
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	<-make(chan struct{}) // Stall the bot until CTRL-C is pressed
}

var channel = make(chan int)
var requestchannel = make(chan int)

type clientIDTransport struct {
	Transport http.RoundTripper
	ClientID  string
}

func (c *clientIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}
	req.Header.Add("X-MAL-CLIENT-ID", c.ClientID)
	return c.Transport.RoundTrip(req)
}
func fetchmal(content string) [3][2]string {
	publicInfoClient := &http.Client{
		// Create client ID from https://myanimelist.net/apiconfig.
		Transport: &clientIDTransport{ClientID: "6191b239cf01f8a3f46d24fd8760c899"},
	}
	ctx := context.Background()
	c := mal.NewClient(publicInfoClient)
	anime, _, _ := c.Anime.List(ctx, content, mal.Limit(3))
	var results [3][2]string
	for b, a := range anime {
		for i := 0; i < 2; i++ {
			if i == 0 {
				results[b][i] = a.Title
			} else if i == 1 {
				results[b][i] = string(rune(a.ID))
			}
		}
	}
	return results
}
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.Content)
	isanime := false
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.ToLower(m.Content) == "fucking terrible bot" {
		_, err := s.ChannelMessageSend(m.ChannelID, "im sad")
		if err != nil {
			return
		}
	}
	if strings.ToLower(m.Content) == "good bot" {
		_, err := s.ChannelMessageSend(m.ChannelID, "im happy !")
		if err != nil {
			return
		}
	}

	select {
	case bin := <-requestchannel:
		//We have request
		if bin == -1 {
			if m.Content == "1" || m.Content == "2" || m.Content == "3" {
				u, _ := strconv.Atoi(m.Content)
				u = u - 1
				channel <- u
			} else {
				if m.Content[0:1] == "." {

					if m.Content[0:2] == ".a" {
						m.Content = m.Content[2:]
						isanime = true
					} else {
						m.Content = m.Content[1:]
					}
					channel <- -1 //Signals to cancel the operation
					if isanime == true {
						go animeroute(s, m)
					} else {
						go x1337route(s, m)
					}
				} else {
					requestchannel <- -1
					return
				}
			}
		}
	default:
		//As there is no request operate is if this is the first experience
		if m.Content[0:1] == "." {
			if m.Content[0:2] == ".a" {
				m.Content = m.Content[2:]
				isanime = true
			} else {
				m.Content = m.Content[1:]
			}
			if isanime == true {
				go animeroute(s, m)
			} else {
				go x1337route(s, m)
			}
		} else {
			return
		}
	}
}
func x1337route(s *discordgo.Session, m *discordgo.MessageCreate) {

	_, err := s.ChannelMessageSend(m.ChannelID, "Ok I will find this thing's magnet link")
	if err != nil {
		return
	}
	bow := surf.NewBrowser()
	fmt.Println("Flag A")
	_ = bow.Open(fmt.Sprintf("https://1337x.to/search/%s/1/", m.Content))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	var results []string
	selector := "body > main > div > div > div > div.box-info-detail.inner-table > div.table-list-wrap > table > tbody"
	bow.Dom().Find(selector).Find("tr").Each(func(count int, s *goquery.Selection) {
		name := s.Find("td.coll-1").Text()
		readd, _ := s.Find("td.coll-1").Find("a.icon").Find("i").Attr("class")
		readd = readd[9:]
		if readd == "video-dual-sound" || readd == "divx" || readd == "hd" || readd == "h264" {
			readd = "Video"
		}
		readd = "(" + readd + ")" + ": " + name
		results = append(results, readd)
	})
	fmt.Println(results)

	if len(results) > 0 {
		if len(results) > 3 {
			for i := 0; i < 3; i++ {
				_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v. %s \n", i, results[i]))
				if err != nil {
					return
				}
			}
		} else {
			x := len(results)
			for i := 0; i < x; i++ {
				_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v. %s \n", i, results[i]))
				if err != nil {
					return
				}
			}
		}
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "No results found")
		if err != nil {
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Pick one")
	if err != nil {
		return
	}

	requestchannel <- -1 //Sends a request saying that we need a 0,1,2
	v := <-channel       //Waits until there is a -1,0,1,2

	if v == -1 { //If there has been a new request then we cancel the current process
		fmt.Println("They made a new process")
		_, err := s.ChannelMessageSend(m.ChannelID, "Ok then")
		if err != nil {
			return
		}
		return
	}

	Newselector := fmt.Sprintf("body > main > div > div > div > div.box-info-detail.inner-table > div.table-list-wrap > table > tbody > tr:nth-child(%v) > td.coll-1.name > a:nth-child(2)", v+1)
	_ = bow.Click(Newselector)

	//We are now in the chosen item's info page

	BodyMainDivDivDiv := "body > main > div > div > div"                                                      //For reuse
	lastworkaround := "ul:nth-child(2) > li:nth-child(1) > span"                                              //for clean
	category := bow.Find(BodyMainDivDivDiv).Find("div.no-top-radius").Find("div").Find(lastworkaround).Text() //We find the category
	if category == "TV" || category == "Movies" {                                                             //If it's a movie or a tv then it will serve a streamable link instead
		bow.Find(BodyMainDivDivDiv).Find("div.no-top-radius").Find("div.clearfix").Find("ul").Find("li").Each(func(count int, selec *goquery.Selection) {
			if selec.Find("a").Text() == "Play now (Stream)" {
				thing2, _ := selec.Find("a").Attr("href")
				_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed.NewGenericEmbed("Your TV/Movie", thing2))
				if err != nil {
					return
				}
			}
		})
	} else {
		adescriptivename := "body > main > div > div > div"
		magnetlink, _ := bow.Find(adescriptivename).Find("div.no-top-radius").Find("div.clearfix").Find("ul").Find("li").Find("a").Attr("href")
		_, err := s.ChannelMessageSend(m.ChannelID, "Magnet Link will be sent to your dms.")
		if err != nil {
			return
		}
		chann, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			fmt.Println("error creating channel:", err)
			_, err := s.ChannelMessageSend(m.ChannelID, "Something went wrong while sending the DM!")
			if err != nil {
				return
			}
			return
		}
		_, err = s.ChannelMessageSend(chann.ID, magnetlink)
		if err != nil {
			fmt.Println("error sending DM message:", err)
			_, err := s.ChannelMessageSend(m.ChannelID, "Failed to send you a DM. "+"Did you disable DM in your privacy settings?")
			if err != nil {
				return
			}
		}
	}

}
func animeroute(s *discordgo.Session, m *discordgo.MessageCreate) {

	var content string
	for count := 0; len(m.Content) > count; count++ {
		if m.Content[count] == ' ' {
			if count == 0 {
			} else {
				content = content + string('-')
			}
		} else {
			content = content + string(m.Content[count])
		}
	}

	var episode string
	var episodeflag bool = false

	if strings.Contains(m.Content, "episode") {
		episodeindex := strings.Index(m.Content, "episode")
		episode = m.Content[episodeindex:]
		fmt.Println(episode)
		episodeflag = true
	}

	//client := &http.Client{}

	type Anime struct {
		Error   bool
		Referer string
	}

	url := fmt.Sprintf("http://localhost:3000/gogoanime/watch/%s", content)
	fmt.Println(url)
	res, err := http.Get(url)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Tell Addison to go fuck himself")
		if err != nil {
			return
		}
		fmt.Println(err)
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(res.Body)
		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			fmt.Println("Error 4")
			log.Fatal(readErr)
		}

		var anime Anime
		jsonErr := json.Unmarshal(body, &anime)
		if jsonErr != nil {
			fmt.Println("Error 5")
			log.Fatal(jsonErr)
		}
		if anime.Error {
			results := fetchmal(content)

			var message string
			for i := 0; i < 3; i++ {
				if results[i][0] == "" {
					return
				}
				message = message + fmt.Sprintf("%d: %s\n", i+1, results[i][0])
			}
			if len(message) == 0 {
				_, err := s.ChannelMessageSend(m.ChannelID, "There is no anime under that name")
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			_, err = s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				log.Fatal(err)
				return
			}
			_, err = s.ChannelMessageSend(m.ChannelID, "Pick one")
			if err != nil {
				log.Fatal(err)
				return
			}
			fmt.Println("B")
			requestchannel <- -1 //Sends a request saying that we need a 0,1,2
			fmt.Println("W")
			v := <-channel //Waits until there is a -1,0,1,2

			fmt.Println("Channel Check succeeded")

			if v == -1 { //If there has been a new request then we cancel the current process
				fmt.Println("They made a new process")
				_, err := s.ChannelMessageSend(m.ChannelID, "Ok then")
				if err != nil {
					return
				}
				return
			}

			fmt.Println("Episode flag")
			if episodeflag {
				url := fmt.Sprintf("http://localhost:3000/gogoanime/watch/%s %s", results[v][0], episode)
				fmt.Println(url)
				var durl string
				for i := 0; i < len(url); i++ {
					if url[i] == ' ' {
						durl = durl + "-"
					} else {
						durl = durl + string(url[i])
					}
				}
				fmt.Println(durl)
				res, err := http.Get(durl)
				if err != nil {
					_, err := s.ChannelMessageSend(m.ChannelID, "Tell Addison to go fuck himself")
					if err != nil {
						return
					}
					fmt.Println(err)
				} else {
					defer func(Body io.ReadCloser) {
						err := Body.Close()
						if err != nil {

						}
					}(res.Body)
					body, readErr := ioutil.ReadAll(res.Body)
					if readErr != nil {
						fmt.Println("Error 4")
						log.Fatal(readErr)
					}
					var newanime Anime
					jsonErr := json.Unmarshal(body, &newanime)
					if jsonErr != nil {
						fmt.Println("Error 5")
						log.Fatal(jsonErr)
					}
					_, err = s.ChannelMessageSend(m.ChannelID, newanime.Referer)
					if err != nil {
						log.Fatal(err)
						return
					}
				}

			} else {
				url := fmt.Sprintf("http://localhost:3000/gogoanime/watch/%s ", results[v][0])
				fmt.Println(url)
				var durl string
				for i := 0; i < len(url); i++ {
					if url[i] == ' ' {
						durl = durl + "-"
					} else {
						durl = durl + string(url[i])
					}
				}
				fmt.Println(durl)
				res, err := http.Get(durl)
				if err != nil {
					_, err := s.ChannelMessageSend(m.ChannelID, "Tell Addison to go fuck himself")
					if err != nil {
						return
					}
					fmt.Println(err)
				} else {
					defer func(Body io.ReadCloser) {
						err := Body.Close()
						if err != nil {

						}
					}(res.Body)
					body, readErr := ioutil.ReadAll(res.Body)
					if readErr != nil {
						fmt.Println("Error 4")
						log.Fatal(readErr)
					}
					var newanime Anime
					jsonErr := json.Unmarshal(body, &newanime)
					if jsonErr != nil {
						fmt.Println("Error 5")
						log.Fatal(jsonErr)
					}
					_, err = s.ChannelMessageSend(m.ChannelID, anime.Referer)
					if err != nil {
						log.Fatal(err)
						return
					}
				}

			}

		} else {
			_, err = s.ChannelMessageSend(m.ChannelID, anime.Referer)
			if err != nil {
				return
			}
		}
	}
}
