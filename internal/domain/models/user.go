package models

type User struct {
	ID          int64
	Username    string
	Email       string
	FIO         string
	PhoneNumber string
	PassHash    []byte
}
