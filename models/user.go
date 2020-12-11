package models

type User struct {
	Id           string   `json:"id"`
	Email        string   `json:"email" validate:"email required"`
	Password     string   `json:"password" validate:"min=8,max=32 required"`
	PersonalData []string `json:"personal_data"`
}
