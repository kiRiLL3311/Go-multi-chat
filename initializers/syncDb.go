package initializers

import "github.com/kiRiLL3311/Go-multi-chat/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}
