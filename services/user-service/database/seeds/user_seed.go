package seeds

import (
	"log"
	"user-service/internal/core/domain/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) {
	// Check if users already exist
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		log.Println("Users already seeded, skipping...")
		return
	}

	// Hash password for test users
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create test users
	users := []model.User{
		{
			Name:       "Super Admin",
			Email:      "superadmin@example.com",
			Password:   string(hashedPassword),
			Phone:      "+628123456789",
			Lat:        "-6.2088",
			Lng:        "106.8456",
			IsVerified: true,
		},
		{
			Name:       "John Customer",
			Email:      "john@example.com",
			Password:   string(hashedPassword),
			Phone:      "+628987654321",
			Lat:        "-6.2000",
			Lng:        "106.8167",
			IsVerified: true,
		},
		{
			Name:       "Jane Customer",
			Email:      "jane@example.com",
			Password:   string(hashedPassword),
			Phone:      "+628112233445",
			Lat:        "-6.1751",
			Lng:        "106.8650",
			IsVerified: true,
		},
		{
			Name:       "Bob Customer",
			Email:      "bob@example.com",
			Password:   string(hashedPassword),
			Phone:      "+628556667778",
			Lat:        "-6.2146",
			Lng:        "106.8451",
			IsVerified: true,
		},
		{
			Name:       "Alice Customer",
			Email:      "alice@example.com",
			Password:   string(hashedPassword),
			Phone:      "+628998877665",
			Lat:        "-6.2088",
			Lng:        "106.8456",
			IsVerified: true,
		},
	}

	// Insert users
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to seed user %s: %v", user.Email, err)
		} else {
			log.Printf("Seeded user: %s (%s)", user.Name, user.Email)
		}
	}

	log.Printf("Successfully seeded %d users", len(users))
}
