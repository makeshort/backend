package request

type UserCreate struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type URL struct {
	Url   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type RefreshToken struct {
	Token string `json:"token"`
}
