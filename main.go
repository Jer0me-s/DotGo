package main

import (
	"flag"
	"fmt"
	"github.com/Clinet/discordgo-embed"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/headzoo/surf.v1"
	"strconv"
)

var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	Token = ""

	dg, err := discordgo.New("Bot " + Token)
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
	<-make(chan struct{})
	dg.Close()
}

var channel = make(chan int)
var requestchannel = make(chan int)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.Content)
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content[0:1] == "." {
		m.Content = m.Content[1:]
	} else {
		return
	}

	select {
	case bin := <-requestchannel:
		//We have request
		if bin == -1 {
			if m.Content == "0" || m.Content == "1" || m.Content == "2" {
				u, _ := strconv.Atoi(m.Content)
				channel <- u
			} else {
				channel <- -1 //Signals to cancel the operation
				go mainroute(s, m)
			}
		}
	default:
		//As there is no request operate is if this is the first experience
		if m.Content == "0" || m.Content == "1" || m.Content == "2" {
			s.ChannelMessageSend(m.ChannelID, "There is nothing to choose from")
			return
		} else {
			go mainroute(s, m)
		}
	}
}

//TODO add anime compatibility (GET BOWEN)

func mainroute(s *discordgo.Session, m *discordgo.MessageCreate) {
	channelID := m.ChannelID
	content := m.Content
	s.ChannelMessageSend(channelID, "Ok I will find this thing's magnet link")
	bow := surf.NewBrowser()
	err := bow.Open(fmt.Sprintf("https://1337x.to/search/%s/1/", content))
	if err != nil {
		panic(err)
	}
	results := []string{}
	selector := "body > main > div > div > div > div.box-info-detail.inner-table > div.table-list-wrap > table > tbody"
	bow.Dom().Find(selector).Find("tr").Each(func(count int, s *goquery.Selection) {
		name := s.Find("td.coll-1").Text()
		results = append(results, name)
	})
	for i := 0; i < 3; i++ {
		s.ChannelMessageSend(channelID, fmt.Sprintf("%v. %s \n", i, results[i])) //TODO add if its a movie tv show etc
	}
	s.ChannelMessageSend(channelID, "Pick one")
	//(<-channel)
	fmt.Println("Flag 1")
	requestchannel <- -1
	fmt.Println("Flag 2")
	//Sends a request saying that we need a 0,1,2
	v := <-channel //Waits until there is a -1,0,1,2

	if v == -1 { //If there has been a new request then we cancel the current process
		fmt.Println("They made a new ")
		s.ChannelMessageSend(channelID, "Stopped the current process")
		return
	}
	Newselector := fmt.Sprintf("body > main > div > div > div > div.box-info-detail.inner-table > div.table-list-wrap > table > tbody > tr:nth-child(%v) > td.coll-1.name > a:nth-child(2)", v+1)
	bow.Click(Newselector)
	frontworkaround := "body > main > div > div > div"
	lastworkaround := "ul:nth-child(2) > li:nth-child(1) > span"
	egg := bow.Find(frontworkaround).Find("div.no-top-radius").Find("div").Find(lastworkaround).Text()
	if egg == "TV" || egg == "Movies" {
		fmt.Println("Goodboy")
		streamselector := "#l3f908a6dc924d6f4ce84fdeea71cb688db46679c"
		bow.Click(streamselector)
		msg := fmt.Sprintf("%s: %s", bow.Title(), bow.Url())
		s.ChannelMessageSendEmbed(channelID, embed.NewGenericEmbed("Your TV/Movie", msg))
	} else {
		adescriptivename := "body > main > div > div > div"
		magnetlink, _ := bow.Find(adescriptivename).Find("div.no-top-radius").Find("div.clearfix").Find("ul").Find("li").Find("a").Attr("href")
		s.ChannelMessageSend(channelID, "Magnet Link will be sent to your dms.")
		chann, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			fmt.Println("error creating channel:", err)
			s.ChannelMessageSend(m.ChannelID, "Something went wrong while sending the DM!")
			return
		}
		_, err = s.ChannelMessageSend(chann.ID, magnetlink)
		if err != nil {
			fmt.Println("error sending DM message:", err)
			s.ChannelMessageSend(m.ChannelID, "Failed to send you a DM. "+"Did you disable DM in your privacy settings?")
		}
	}

}
