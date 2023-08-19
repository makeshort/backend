package request

type UserCreate struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLogIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type URL struct {
	Url   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}
