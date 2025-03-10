package slacker

import (
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// MessageEvent contains details common to message based events, including the
// raw event as returned from Slack along with the corresponding event type.
// The struct should be kept minimal and only include data that is commonly
// used to prevent frequent type assertions when evaluating the event.
type MessageEvent struct {
	// Channel ID where the message was sent
	ChannelID string

	// Channel contains information about the channel
	Channel *slack.Channel

	// User ID of the sender
	UserID string

	// UserProfile contains all the information details of a given user
	UserProfile *slack.UserProfile

	// Text is the unalterted text of the message, as returned by Slack
	Text string

	// TimeStamp is the message timestamp. For events that do not support
	// threading (eg. slash commands) this will be unset.
	// will be left unset.
	TimeStamp string

	// ThreadTimeStamp is the message thread timestamp. For events that do not
	// support threading (eg. slash commands) this will be unset.
	ThreadTimeStamp string

	// Data is the raw event data returned from slack. Using Type, you can assert
	// this into a slackevents *Event struct.
	Data interface{}

	// Type is the type of the event, as returned by Slack. For instance,
	// `app_mention` or `message`
	Type string

	// BotID of the bot that sent this message. If a bot did not send this
	// message, this will be an empty string.
	BotID string
}

// IsThread indicates if a message event took place in a thread.
func (e *MessageEvent) IsThread() bool {
	if e.ThreadTimeStamp == "" || e.ThreadTimeStamp == e.TimeStamp {
		return false
	}
	return true
}

// IsBot indicates if the message was sent by a bot
func (e *MessageEvent) IsBot() bool {
	return e.BotID != ""
}

// NewMessageEvent creates a new message event structure
func NewMessageEvent(slacker *Slacker, event interface{}, req *socketmode.Request) *MessageEvent {
	var messageEvent *MessageEvent

	switch ev := event.(type) {
	case *slackevents.MessageEvent:
		messageEvent = &MessageEvent{
			ChannelID:       ev.Channel,
			Channel:         getChannel(slacker, ev.Channel),
			UserID:          ev.User,
			UserProfile:     getUserProfile(slacker, ev.User),
			Text:            ev.Text,
			Data:            event,
			Type:            ev.Type,
			TimeStamp:       ev.TimeStamp,
			ThreadTimeStamp: ev.ThreadTimeStamp,
			BotID:           ev.BotID,
		}
	case *slackevents.AppMentionEvent:
		messageEvent = &MessageEvent{
			ChannelID:       ev.Channel,
			Channel:         getChannel(slacker, ev.Channel),
			UserID:          ev.User,
			UserProfile:     getUserProfile(slacker, ev.User),
			Text:            ev.Text,
			Data:            event,
			Type:            ev.Type,
			TimeStamp:       ev.TimeStamp,
			ThreadTimeStamp: ev.ThreadTimeStamp,
			BotID:           ev.BotID,
		}
	case *slack.SlashCommand:
		messageEvent = &MessageEvent{
			ChannelID:   ev.ChannelID,
			Channel:     getChannel(slacker, ev.ChannelID),
			UserID:      ev.UserID,
			UserProfile: getUserProfile(slacker, ev.UserID),
			Text:        fmt.Sprintf("%s %s", ev.Command[1:], ev.Text),
			Data:        req,
			Type:        req.Type,
		}
	default:
		return nil
	}

	return messageEvent
}

func getChannel(slacker *Slacker, channelID string) *slack.Channel {
	if len(channelID) == 0 {
		return nil
	}

	channel, err := slacker.apiClient.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID:         channelID,
		IncludeLocale:     false,
		IncludeNumMembers: false})
	if err != nil {
		slacker.logf("unable to get channel info for %s: %v\n", channelID, err)
		return nil
	}
	return channel
}

func getUserProfile(slacker *Slacker, userID string) *slack.UserProfile {
	if len(userID) == 0 {
		return nil
	}

	user, err := slacker.apiClient.GetUserInfo(userID)
	if err != nil {
		slacker.logf("unable to get user info for %s: %v\n", userID, err)
		return nil
	}
	return &user.Profile
}
