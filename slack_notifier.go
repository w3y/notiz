package notiz

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"github.com/w3y/notiz/utils"
	"net"
	"net/http"
	"sync"
	"time"
)

const _pressureThreshold = 800

var (
	netClient = &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	lock          sync.RWMutex
	notifyPusher  *slackNotifierPusher
	membersPicked = make(map[string]string)
)

func NewSlackNotifyPusher() *slackNotifierPusher {
	pusher := &slackNotifierPusher{
		pushChan: make(chan pusherMessage, 1000),
	}

	go func() {
		for {
			select {
			case <-closeChan:
				return
			case msg := <-pusher.pushChan:
				buf := bytes.NewBuffer(msg.message)
				_, err := netClient.Post(msg.webHook, "application/json", buf)
				if err != nil {
					//TODO: print error log
				}
			}
		}
	}()

	return pusher
}

type pusherMessage struct {
	webHook string
	message []byte
}

type slackNotifierPusher struct {
	pushChan chan pusherMessage
}

type SlackMessage struct {
	TaggingMsg  string   `json:"pretext"`
	Color       string   `json:"color"`
	App         string   `json:"author_name"`
	AppLink     string   `json:"author_link"`
	NotiContent string   `json:"title"`
	RefLink     string   `json:"title_link"`
	NotiDetail  string   `json:"text"`
	Env         string   `json:"footer"`
	Ts          int64    `json:"ts"`
	members     []string `json:"-"`
}

func init() {
	lock.Lock()
	defer lock.Unlock()

	if notifyPusher == nil {
		notifyPusher = NewSlackNotifyPusher()
	}
}

func (m *SlackMessage) Opts() map[string]string {
	return map[string]string{}
}

func (m *SlackMessage) Hash() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", m.Env, m.App, m.NotiContent, utils.GetMD5Hash(m.NotiDetail), m.RefLink)
}

type SlackNotifier struct {
	webHook    string
	notifyChan chan []byte
	lazyNoti   bool
	members    []string
}

func NewSlackNotifier() *SlackNotifier {
	return NewSlackNotifierCustom(viper.GetString("notifier.slack.webhook"))
}

func NewSlackCriticalNotifier() *SlackNotifier {
	return NewSlackNotifierCustom(viper.GetString("notifier.critical.slack.webhook"))
}

func NewSlackNotifierCustomWithMembers(customWebhook string, members []string) *SlackNotifier {
	slackNotifier := &SlackNotifier{
		webHook:  customWebhook,
		lazyNoti: true,
		members:  members,
	}

	return slackNotifier
}

func NewSlackNotifierCustom(customWebhook string) *SlackNotifier {
	return NewSlackNotifierCustomWithMembers(customWebhook, []string{})
}
