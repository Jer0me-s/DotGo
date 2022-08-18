package main

import (
	"flag"
	"fmt"
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
	Token = "OTg4NTg3NzE1Mzc1MjcxOTY2.GV-hRp.etlZTbHlAz0RuiMlzWbbMGwxNl5aiL-tY8z0tA"

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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.Content)
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "0" || m.Content == "1" || m.Content == "2" {
		u, _ := strconv.Atoi(m.Content)
		channel <- u

	} else {
		go mainroute(s, m)
	}
}

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
		s.ChannelMessageSend(channelID, fmt.Sprintf("%v. %s \n", i, results[i]))
	}
	s.ChannelMessageSend(channelID, "Pick one fucker")
	//(<-channel)
	Newselector := fmt.Sprintf("body > main > div > div > div > div.box-info-detail.inner-table > div.table-list-wrap > table > tbody > tr:nth-child(%v) > td.coll-1.name > a:nth-child(2)", <-channel+1)
	bow.Click(Newselector)

	fmt.Println(bow.Title())
}
