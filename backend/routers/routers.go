package routers

import (
	jwt_service "LinkShorter/jwt"
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

	r.POST("/api/registration", func(c *gin.Context) {
		var req CreateRegistrationRequest
		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"Неправильный запрос"},
			})
			return
		}

		if len(req.Username) > 30 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"Максимальная длина имени пользователя 30 символов"},
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
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		if usernameExist {
			errs = append(errs, "Данное имя пользователя уже используется")
		}

		emailExist, err = services.IsEmailInDatabase(db, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
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
				"errors": []string{"Ошибка сервера. Попробуйте позже"},
			})
			return
		}

		err = services.RegisterUser(db, username, email, string(hashedPassword))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		err = services.SendEmail(db, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка отправки письма. Попробуйте позже"},
			})
			return
		}

		var userID int
		userID, err = services.GetUserID(db, username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		var token string
		token, err = jwt_service.GenerateJWT(userID, false)

		c.SetCookie("token", token, 30*24*3600, "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{})
	})

	r.POST("/api/login", func(c *gin.Context) {
		var req CreateLoginRequest
		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"Неправильный запрос"},
			})
			return
		}

		var loginInput, password string
		loginInput = strings.TrimSpace(req.LoginInput)
		password = strings.TrimSpace(req.Password)

		var isExist bool
		isExist, err = services.IsLoginValid(db, loginInput, password)

		if !isExist && err == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Неправильный логин или пароль"},
			})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		var userID int
		userID, err = services.GetUserID(db, loginInput)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		var isVerified bool
		isVerified, err = services.GetUserVerified(db, loginInput)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		var token string
		token, err = jwt_service.GenerateJWT(userID, isVerified)

		c.SetCookie("token", token, 30*24*3600, "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{})
	})
}
