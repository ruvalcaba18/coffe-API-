package database

import (
	"coffeebase-api/internal/auth"
	"database/sql"
	"log"
)

// SeedDatabase inserts initial default users, like super admins.
func SeedDatabase(db *sql.DB) error {
	superAdmins := []struct {
		Username string
		Email    string
		Password string
		Role     string
	}{
		{
			Username: "jael",
			Email:    "jael.ruvalcaba@uabc.edu.mx",
			Password: "user123!",
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
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", admin.Email).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			hashed, err := auth.HashPassword(admin.Password)
			if err != nil {
				return err
			}

			_, err = db.Exec(`
				INSERT INTO users (username, email, password, language, role) 
				VALUES ($1, $2, $3, 'es', $4)`,
				admin.Username, admin.Email, hashed, admin.Role,
			)
			if err != nil {
				return err
			}
			log.Printf("✅ Seeded super admin: %s (%s)", admin.Username, admin.Email)
		}
	}

	return nil
}
