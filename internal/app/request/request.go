package request

type UserCreate struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserUpdate struct {
	Email      string `json:"email,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	TelegramID string `json:"telegram_id,omitempty"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type URL struct {
	Url   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type UrlUpdate struct {
	Url   string `json:"url,omitempty"`
	Alias string `json:"alias,omitempty"`
}
