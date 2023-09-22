package users

type UserLogin struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsDriver bool   `json:"isDriver"`
	Token    Token  `json:"token"`
}

type Token struct {
	Value     *string `json:"value"`
	ExpiresIn int64   `json:"expiresIn"`
}
