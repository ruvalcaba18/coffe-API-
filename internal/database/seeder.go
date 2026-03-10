package database

import (
	"coffeebase-api/internal/auth"
	"database/sql"
	"log"
	stringManipulation "strings"
)

func SeedDatabase(db *sql.DB) error {
	superAdmins := []struct {
		Username string
		Email    string
		Password string
		Role     string
	}{
		{
			Username: "Jael Admin",
			Email:    "jael.ruvalcaba@uabc.edu.mx",
			Password: "User123!",
			Role:     "superadmin",
		},
		{
			Username: "ian",
			Email:    "ian@gmail.com",
			Password: "user123!",
			Role:     "superadmin",
		},
	}

	for _, admin := range superAdmins {
		admin.Email = stringManipulation.ToLower(stringManipulation.TrimSpace(admin.Email))
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))", admin.Email).Scan(&exists)
		if err != nil {
			return err
		}

		hashed, err := auth.HashPassword(admin.Password)
		if err != nil {
			return err
		}

		if !exists {
			_, err = db.Exec(`
				INSERT INTO users (username, email, password, language, role) 
				VALUES ($1, $2, $3, 'es', $4)`,
				admin.Username, admin.Email, hashed, admin.Role,
			)
			if err != nil {
				return err
			}
			log.Printf("✅ Seeded super admin: %s (%s)", admin.Username, admin.Email)
		} else {
			_, err = db.Exec(`
				UPDATE users SET role = $1, password = $2, username = $3
				WHERE email = $4`,
				admin.Role, hashed, admin.Username, admin.Email,
			)
			if err != nil {
				return err
			}
			log.Printf("🔄 Updated super admin: %s (%s)", admin.Username, admin.Email)
		}
	}

	return nil
}
