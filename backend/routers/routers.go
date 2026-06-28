package routers

import (
	"LinkShorter/services"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type CreateLinkRequest struct {
	URL string `json:"url"`
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

func SetupRoutes(r *gin.Engine, db *pgxpool.Pool) {

	r.POST("/create-link", func(c *gin.Context) {
		var req CreateLinkRequest
		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		shortLink, err := services.GenerateLink(db, req.URL)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. Не удалось создать ссылку",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"short-link": shortLink,
		})
	})

	r.GET("/:code", func(c *gin.Context) {
		code := c.Param("code")
		originalURL, err := services.ReturnOriginalLink(db, code)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Ссылки не существует",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка сервера",
			})
			return
		}

		c.Redirect(http.StatusFound, originalURL)
	})

	// НАПИСАТЬ ОТПРАВКУ ПИСЬМА
	r.POST("/api/registration", func(c *gin.Context) {
		var req CreateRegistrationRequest
		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": "Неправильный запрос",
			})
			return
		}

		if len(req.Username) > 30 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": "Максимальная длина имени пользователя 30 символов",
			})
			return
		}

		// Приведение данных в адекватный вид
		var username, email, password string
		username = strings.TrimSpace(req.Username)
		email = strings.TrimSpace(req.Email)
		password = strings.TrimSpace(req.Password)

		var usernameExist bool
		var emailExist bool

		var errs []string

		usernameExist, err = services.IsUsernameInDatabase(db, username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": "Ошибка базы данных. Попробуйте позже",
			})
			return
		}

		if usernameExist {
			errs = append(errs, "Данное имя пользователя уже используется")
		}

		emailExist, err = services.IsEmailInDatabase(db, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": "Ошибка базы данных. Попробуйте позже",
			})
			return
		}

		if emailExist {
			errs = append(errs, "Данная почта уже используется")
		}

		if len(errs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}

		// Хэширование пароля
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": "Ошибка сервера. Попробуйте позже",
			})
			return
		}

		err = services.RegisterUser(db, username, email, string(hashedPassword))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": "Ошибка базы данных. Попробуйте позже",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{})
	})

	r.POST("/api/login", func(c *gin.Context) {
		var req CreateLoginRequest
		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

	})
}
