package ChatService

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
)

const (
	// used to track clients that used chat. mainly for listing clients in the /clients api, in real world chat app
	// such client list should be separated into client management module.
	clientsKey       = "clients"
	clientChannelFmt = "client:%d:channels"
	ChannelsKey      = "channels"
)

type Client struct {
	id              int
	channelsHandler *redis.PubSub

	stopListenerChan chan struct{}
	listening        bool

	MessageChan chan redis.Message
}

// Connect connect client to client channels on redis
func Connect(rdb *redis.Client, id int) (*Client, error) {
	if _, err := rdb.SAdd(clientsKey, id).Result(); err != nil {
		return nil, err
	}

	c := &Client{
		id:               id,
		stopListenerChan: make(chan struct{}),
		MessageChan:      make(chan redis.Message),
	}

	if err := c.connect(rdb); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Subscribe(rdb *redis.Client, channel string) error {

	clientChannelsKey := fmt.Sprintf(clientChannelFmt, c.id)

	if rdb.SIsMember(clientChannelsKey, channel).Val() {
		return nil
	}
	if err := rdb.SAdd(clientChannelsKey, channel).Err(); err != nil {
		return err
	}

	return c.connect(rdb)
}

func (c *Client) Unsubscribe(rdb *redis.Client, channel string) error {

	clientChannelsKey := fmt.Sprintf(clientChannelFmt, c.id)

	if !rdb.SIsMember(clientChannelsKey, channel).Val() {
		return nil
	}
	if err := rdb.SRem(clientChannelsKey, channel).Err(); err != nil {
		return err
	}

	return c.connect(rdb)
}

func (c *Client) connect(rdb *redis.Client) error {

	var c0 []string

	c1, err := rdb.SMembers(ChannelsKey).Result()
	if err != nil {
		return err
	}
	c0 = append(c0, c1...)

	// get all client channels (from DB) and start subscribe
	c2, err := rdb.SMembers(fmt.Sprintf(clientChannelFmt, c.id)).Result()
	if err != nil {
		return err
	}
	c0 = append(c0, c2...)

	if len(c0) == 0 {
		fmt.Println("no channels to connect to for client: ", c.id)
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
		fmt.Println("starting the listener for client:", c.id, "on channels:", channels)
		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					return
				}
				c.MessageChan <- *msg

			case <-c.stopListenerChan:
				fmt.Println("stopping the listener for client:", c.id)
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

func List(rdb *redis.Client) ([]string, error) {
	return rdb.SMembers(clientsKey).Result()
}

func GetChannels(rdb *redis.Client, clientname string) ([]string, error) {

	if !rdb.SIsMember(clientsKey, clientname).Val() {
		return nil, errors.New("client not exists")
	}

	var c []string

	c1, err := rdb.SMembers(ChannelsKey).Result()
	if err != nil {
		return nil, err
	}
	c = append(c, c1...)

	// get all client channels (from DB) and start subscribe
	c2, err := rdb.SMembers(fmt.Sprintf(clientChannelFmt, clientname)).Result()
	if err != nil {
		return nil, err
	}
	c = append(c, c2...)

	return c, nil
}
