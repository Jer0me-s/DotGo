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
	Token = "OTg4NTg3NzE1Mzc1MjcxOTY2.Gu1GuW.sDtRY6L_ewDUsiqKZnPRdvbCFdqtMh-RFVEHaU"

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
	if m.Content == "Fucking cunt" {
		s.ChannelMessageSend(m.ChannelID, "im sad")
	}
	if m.Content == "Good bot" {
		s.ChannelMessageSend(m.ChannelID, "im happy !")
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

	s.ChannelMessageSend(m.ChannelID, "Ok I will find this thing's magnet link")
	bow := surf.NewBrowser()
	fmt.Println("Flag A")
	err := bow.Open(fmt.Sprintf("https://1337x.to/search/%s/1/", m.Content))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	var results []string
	selector := "body > main > div > div > div > div.box-info-detail.inner-table > div.table-list-wrap > table > tbody"
	bow.Dom().Find(selector).Find("tr").Each(func(count int, s *goquery.Selection) {
		name := s.Find("td.coll-1").Text()
		results = append(results, name)
	})
	fmt.Println(results)
	for i := 0; i < 3; i++ {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v. %s \n", i, results[i])) //TODO add if its a movie tv show etc
	}
	s.ChannelMessageSend(m.ChannelID, "Pick one")

	requestchannel <- -1 //Sends a request saying that we need a 0,1,2
	v := <-channel       //Waits until there is a -1,0,1,2

	if v == -1 { //If there has been a new request then we cancel the current process
		fmt.Println("They made a new process")
		s.ChannelMessageSend(m.ChannelID, "Ok then")
		return
	}

	Newselector := fmt.Sprintf("body > main > div > div > div > div.box-info-detail.inner-table > div.table-list-wrap > table > tbody > tr:nth-child(%v) > td.coll-1.name > a:nth-child(2)", v+1)
	bow.Click(Newselector)

	//We are now in the chosen item's info page

	BodyMainDivDivDiv := "body > main > div > div > div"                                                      //For reuse
	lastworkaround := "ul:nth-child(2) > li:nth-child(1) > span"                                              //for clean
	category := bow.Find(BodyMainDivDivDiv).Find("div.no-top-radius").Find("div").Find(lastworkaround).Text() //We find the category
	if category == "TV" || category == "Movies" {                                                             //If it's a movie or a tv then it will serve a streamable link instead
		bow.Find(BodyMainDivDivDiv).Find("div.no-top-radius").Find("div.clearfix").Find("ul").Find("li").Each(func(count int, selec *goquery.Selection) {
			if selec.Find("a").Text() == "Play now (Stream)" {
				thing2, _ := selec.Find("a").Attr("href")
				s.ChannelMessageSendEmbed(m.ChannelID, embed.NewGenericEmbed("Your TV/Movie", thing2))
			}
		})
	} else {
		adescriptivename := "body > main > div > div > div"
		magnetlink, _ := bow.Find(adescriptivename).Find("div.no-top-radius").Find("div.clearfix").Find("ul").Find("li").Find("a").Attr("href")
		s.ChannelMessageSend(m.ChannelID, "Magnet Link will be sent to your dms.")
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
