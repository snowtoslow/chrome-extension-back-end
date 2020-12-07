package models

type User struct {
	Email string `json:"email"`
	Password string `json:"password"`
	PersonalData []string `json:"personal_data"`
}
