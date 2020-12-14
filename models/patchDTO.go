package models

type PatchDTO struct {
	PersonalData []string `json:"personal_data,omitempty"`
	Email        string   `json:"email"`
}
