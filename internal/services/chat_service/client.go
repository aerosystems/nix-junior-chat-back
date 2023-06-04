package ChatService

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/go-redis/redis/v7"
)

const (
	clientsKey  = "clients"
	channelsKey = "channels"
)

type Client struct {
	User            *models.User
	channelsHandler *redis.PubSub

	stopListenerChan chan struct{}
	listening        bool

	MessageChan chan redis.Message
}

// Connect connect client to client channels on redis
func Connect(rdb *redis.Client, user *models.User) (*Client, error) {
	if _, err := rdb.SAdd(clientsKey, user.ID).Result(); err != nil {
		return nil, err
	}

	c := &Client{
		User:             user,
		stopListenerChan: make(chan struct{}),
		MessageChan:      make(chan redis.Message),
	}

	if err := c.connect(rdb); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) connect(rdb *redis.Client) error {

	var c0 []string

	c1, err := rdb.SMembers(channelsKey).Result()
	if err != nil {
		return err
	}
	c0 = append(c0, c1...)

	for _, chat := range c.User.Chats {
		c0 = append(c0, fmt.Sprintf("%d", chat.ID))
	}

	if len(c0) == 0 {
		fmt.Println("no channels to connect to for client: ", c.User.ID)
		return nil
	}

	if c.channelsHandler != nil {
		if err := c.channelsHandler.Unsubscribe(); err != nil {
			return err
		}
		if err := c.channelsHandler.Close(); err != nil {
			return err
		}
	}
	if c.listening {
		c.stopListenerChan <- struct{}{}
	}

	return c.doConnect(rdb, c0...)
}

func (c *Client) doConnect(rdb *redis.Client, channels ...string) error {
	// subscribe all channels in one request
	pubSub := rdb.Subscribe(channels...)
	// keep channel handler to be used in unsubscribe
	c.channelsHandler = pubSub

	// The Listener
	go func() {
		c.listening = true
		fmt.Println("starting the listener for client:", c.User.ID, "on channels:", channels)
		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					return
				}
				c.MessageChan <- *msg

			case <-c.stopListenerChan:
				fmt.Println("stopping the listener for client:", c.User.ID)
				return
			}
		}
	}()
	return nil
}

func (c *Client) Disconnect() error {
	if c.channelsHandler != nil {
		if err := c.channelsHandler.Unsubscribe(); err != nil {
			return err
		}
		if err := c.channelsHandler.Close(); err != nil {
			return err
		}
	}
	if c.listening {
		c.stopListenerChan <- struct{}{}
	}

	close(c.MessageChan)

	return nil
}

func Chat(rdb *redis.Client, channel string, content string) error {
	return rdb.Publish(channel, content).Err()
}
