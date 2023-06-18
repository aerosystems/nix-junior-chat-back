package ChatService

import (
	"context"
	"firebase.google.com/go/messaging"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"log"
)

func (cs *ChatService) SendNotification(message *models.Message) {
	chat, err := cs.chatRepo.FindByID(message.ChatID)
	if err != nil {
		log.Println("error getting chat:", err)
		return
	}

	for _, c := range chat.Users {
		if c.ID == message.SenderID {
			continue
		}
		user, err := cs.userRepo.FindByID(c.ID)
		if err != nil {
			log.Println("error getting user:", err)
			continue
		}
		if len(user.Devices) > 0 {
			ctx := context.Background()
			client, err := cs.firebaseApp.Messaging(ctx)
			if err != nil {
				log.Printf("error getting Messaging client: %v\n\n", err)
				continue
			} else {
				for _, d := range user.Devices {
					message := &messaging.Message{
						Notification: &messaging.Notification{
							Title: fmt.Sprintf("Message from %s", message.Sender.Username),
							Body:  message.Content,
						},
						Token: d.Token,
					}
					response, err := client.Send(ctx, message)
					if err != nil {
						log.Println("error sending message:", err)
						continue
					}
					fmt.Println("successfully sent message:", response)
				}
			}
		}

	}

}
