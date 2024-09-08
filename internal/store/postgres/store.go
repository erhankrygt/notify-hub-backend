package postgrestore

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	envvars "notify-hub-backend/configs/env-vars"
)

// Message represents the message model.
type Message struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Recipient string `gorm:"not null" json:"recipient"`
	Content   string `gorm:"not null" json:"content"`
	Sent      bool   `gorm:"default:false" json:"sent"`
}

// Store interface defines the methods to interact with the database.
type Store interface {
	FetchMessages(ctx context.Context, sent bool, limit int) ([]Message, error)
	UpdateMessageStatusToSent(ctx context.Context, id int64) error
	InsertDummyMessages(ctx context.Context) error
	Close() error
}

type store struct {
	db *gorm.DB
}

func NewStore(cfg envvars.Postgres) (Store, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	if err := db.AutoMigrate(&Message{}); err != nil {
		return nil, fmt.Errorf("failed to migrate the Message model: %w", err)
	}

	return &store{db: db}, nil
}

// FetchMessages retrieves messages based on their sent status and applies a limit.
func (s *store) FetchMessages(ctx context.Context, sent bool, limit int) ([]Message, error) {
	var messages []Message
	if err := s.db.WithContext(ctx).Where("sent = ?", sent).Order("id ASC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return messages, nil
}

// UpdateMessageStatusToSent updates the sent status of a message to true based on its ID.
func (s *store) UpdateMessageStatusToSent(ctx context.Context, id int64) error {
	if err := s.db.WithContext(ctx).Model(&Message{}).Where("id = ?", id).Update("sent", true).Error; err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

// InsertDummyMessages deletes all existing messages and inserts dummy messages numbered from 1 to 10 into the Message table.
func (s *store) InsertDummyMessages(ctx context.Context) error {
	// Delete all existing records in the Message table
	if err := s.db.WithContext(ctx).Where("1 = 1").Delete(&Message{}).Error; err != nil {
		return fmt.Errorf("failed to delete existing messages: %w", err)
	}

	// Insert dummy messages numbered from 1 to 10
	for i := 1; i <= 10; i++ {
		message := Message{
			Recipient: fmt.Sprintf("532500808%d", i),
			Content:   fmt.Sprintf("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque sit amet sem nec nisl facilisis pretium. Nunc aliquet justo euismod urna, in fermentum eros accumsan. This is message number %d", i),
			Sent:      false,
		}

		if err := s.db.WithContext(ctx).Create(&message).Error; err != nil {
			return fmt.Errorf("failed to insert dummy message %d: %w", i, err)
		}
	}

	return nil
}

// Close closes the GORM database connection.
func (s *store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to retrieve the generic DB object from GORM: %w", err)
	}

	return sqlDB.Close()
}
