package routers

/*
Файл types.go содержит структуры для биндинга JSON
из тела входящих запросов (request body) на роутах
проекта. Используются в auth.go, links.go
и reset_password.go через ShouldBindJSON
*/

type CreateLinkRequest struct {
	URL string `json:"url"`
}

type CreateCustomLinkRequest struct {
	URL    string `json:"url"`
	Custom string `json:"custom"`
}

type CreateRegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateLoginRequest struct {
	LoginInput string `json:"loginInput"`
	Password   string `json:"password"`
}

type ResendEmail struct {
	Email string `json:"email"`
}

type CheckCode struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResetPassword struct {
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
	ResetToken      string `json:"reset_token"`
}

type AddTag struct {
	LinkID int    `json:"id"`
	Tag    string `json:"tag"`
}
