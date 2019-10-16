package main

type CredentialsInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Image    string `json:"image"`
}

type Credentials struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Image    string `json:"image"`
}
