package model

// User - data for user DataBase
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Avatar   string `json:"image"`
}
