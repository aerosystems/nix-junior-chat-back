package models

type Message struct {
	ID        int    `json:"id" gorm:"primaryKey" example:"1"`
	Type      string `json:"type" example:"message"` // text, error, system
	ChatID    int    `json:"chatId" gorm:"foreignKey:ChatID" example:"1"`
	Content   string `json:"content" example:"bla-bla-bla"`
	SenderID  int    `json:"senderId" gorm:"foreignKey:SenderID" example:"2"`
	Sender    User   `json:"sender"`
	isRead    bool   `json:"isRead" example:"false"`
	CreatedAt int64  `json:"createdAt" example:"1620000000"`
}

type MessageRepository interface {
	FindByID(id int) (*Message, error)
	FindAll() (*[]Message, error)
	Create(message *Message) error
	GetMessages(senderID, recipientID, from, limit int) (*[]Message, error)
}
