package main

import (
	"io/ioutil"
	"encoding/json"
	"github.com/nlopes/slack"
	"log"
	"strings"
	"fmt"
)


type User struct {
	Info   slack.User
	Rating int
}

type Users struct {
	Names []string `json:"users"`
}

type Channels struct {
	Names []string `json:"channels"`
}
type Token struct {
	Token string   `json:"token"`
}

type Message struct {
	ChannelId string
	Timestamp string
	Payload   string
	Rating    int
	User      User
}

type BotCentral struct {
	Channel *slack.Channel
	Event   *slack.MessageEvent
	UserId  string
}

type AttachmentChannel struct {
	Channel      *slack.Channel
	Attachment   *slack.Attachment
	DisplayTitle string
}

type Messages []Message


type ActiveUsers []User

var (
	api *slack.Client
	botKey Token
	whitelistedUsers Users
	whitelistedChannels Channels
	activeUsers ActiveUsers
	userMessages Messages
	botId string
	botCommandChannel chan *BotCentral
	botReplyChannel chan AttachmentChannel
)

func checkValidUser(user string ) bool{
	for _, b := range whitelistedUsers.Names {
		if b == user {
			return true
		}
	}
	return false

}

func checkValidChannel(channel string ) bool{
	for _, b := range whitelistedChannels.Names {
		if b == channel {
			return true
		}}
return false

}

func handleBotCommands(c chan AttachmentChannel) {
	sre_commands := map[string]string{
		"help": "Gives list of commands that are supported with this bot",
	}

	var attachmentChannel AttachmentChannel

	for {
		botChannel := <-botCommandChannel
		attachmentChannel.Channel = botChannel.Channel
		commandArray := strings.Fields(botChannel.Event.Text)
		user1, _ := api.GetUserInfo(botChannel.Event.User)
		channel1, _ := api.GetChannelInfo(botChannel.Event.Channel)
		display := ""
		if !checkValidChannel(channel1.Name) {
			display = "channel"
		}
		if !checkValidUser(user1.Name) {
			display = "user"

		}
		fmt.Println(display)
		if (len(display) > 0) {
			attachmentChannel.DisplayTitle = fmt.Sprintf("Sorry this %s is not whitelisted to execute sre bot commands.", display)
			attachment := &slack.Attachment{
			}
			attachmentChannel.Attachment=attachment
			c <- attachmentChannel
		} else {
			switch commandArray[1] {
			case "help":
				fields := make([]slack.AttachmentField, 0)
				for k, v := range sre_commands {
					fields = append(fields, slack.AttachmentField{
						Title: "{sre} " + k,
						Value: v,
					})
				}

				attachment := &slack.Attachment{
					Pretext: "SRE Commands List",
					Color:   "#61FF33",
					Fields:  fields,
				}
				attachmentChannel.DisplayTitle =""
				attachmentChannel.Attachment = attachment
				c <- attachmentChannel

			}
		}
	}
}
func handleBotReply() {
	for {
		ac := <-botReplyChannel
		params := slack.PostMessageParameters{}
		params.AsUser = true
		params.Attachments = []slack.Attachment{*ac.Attachment}
		_, _, errPostMessage := api.PostMessage(ac.Channel.Name, ac.DisplayTitle, params)
		if errPostMessage != nil {
			log.Fatal(errPostMessage)
		}
	}
}
func init() {
	file, err := ioutil.ReadFile("files/bot_token.json")

	if err != nil {
		log.Fatal("File doesn't exist")
	}

	if err := json.Unmarshal(file, &botKey); err != nil {
		log.Fatal("Cannot parse bot_token.json")
	}

	userFile, err := ioutil.ReadFile("files/users_whitelist.json")

	if err != nil {
		log.Fatal("File doesn't exist")
	}

	if err := json.Unmarshal(userFile, &whitelistedUsers); err != nil {
		log.Fatal("Cannot parse users_whitelist.json")
	}

	channelsFile, err := ioutil.ReadFile("files/channels_whitelist.json")

	if err != nil {
		log.Fatal("File doesn't exist")
	}

	if err := json.Unmarshal(channelsFile, &whitelistedChannels); err != nil {
		log.Fatal("Cannot parse channels_whitelist.json")
	}
}

func main() {
	api = slack.New(botKey.Token) // Create api client with bot token

	rtm := api.NewRTM()

	botCommandChannel = make(chan *BotCentral)
	botReplyChannel = make(chan AttachmentChannel)

	userMessages = make(Messages, 0)

	go rtm.ManageConnection()
	go handleBotCommands(botReplyChannel)
	go handleBotReply()

	for {
		select {

		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {

			case *slack.MessageEvent:
				channelInfo, err := api.GetChannelInfo(ev.Channel)
				if err != nil {
					log.Fatalln(err)
				}
				/*user1, _ := api.GetUserInfo(ev.User)
				channel1,_:= api.GetChannelInfo(ev.Channel)
				fmt.Println(user1.Name);
				fmt.Println(channel1.Name);
				fmt.Println(ev.Text); */

				botCentral := &BotCentral{
					Channel: channelInfo,
					Event: ev,
					UserId: ev.User,
				}
				//&& strings.HasPrefix(ev.Text, "<@" + botId + ">"
				if (ev.Type == "message" && strings.HasPrefix(ev.Text, "<bot>")) {
					botCommandChannel <- botCentral
				}


			default:
				// Ignore other events..
				//fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}