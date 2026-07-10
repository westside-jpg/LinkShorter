package routers

import (
	jwt_service "LinkShorter/jwt"
	"LinkShorter/models"
	"LinkShorter/services"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	qrcode "github.com/skip2/go-qrcode"
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

		// Создание ссылки для незарегистрированного пользователя
		tokenString, err := c.Cookie("token")
		if err != nil {
			shortLink, err := services.GenerateLink(db, req.URL, 0)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ошибка базы данных. Не удалось создать ссылку",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"short-link": shortLink,
			})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Сессия истекла. Войдите в аккаунт",
			})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := int(claims["user_id"].(float64))

		shortLink, err := services.GenerateLink(db, req.URL, userID)
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

		err = services.IncreaseLinkViews(db, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка подсчета статистики",
			})
			return
		}

		c.Redirect(http.StatusFound, originalURL)
	})

	r.GET("/my-links", func(c *gin.Context) {
		userID, err := services.GetUserIdFromJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Ошибка токена. Попробуйте перезайти в аккаунт",
			})
			return
		}

		search := c.DefaultQuery("search", "")
		sortBy := c.DefaultQuery("sort", "date")
		orderBy := c.DefaultQuery("order", "desc")
		filterPeriod := c.DefaultQuery("period", "all")
		filterViews := c.DefaultQuery("views", "0+")
		filterTags := c.DefaultQuery("tags", "all")

		links, err := services.UserLinks(db, userID, search, sortBy, orderBy, filterPeriod, filterViews, filterTags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Ошибка базы данных",
			})
			fmt.Println(err)
			return
		}

		if len(links) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"results": []models.Link{},
				"message": "Ссылок не найдено",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"results": links})
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

		// Приведение данных в адекватный вид
		var username, email, password string
		username = strings.TrimSpace(req.Username)
		email = strings.TrimSpace(req.Email)
		password = strings.TrimSpace(req.Password)

		errs := services.ValidateRegistration(username, email, password)
		if len(errs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}

		var usernameExist bool
		var emailExist bool

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

		var userID int
		userID, err = services.GetUserID(db, username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		var token string
		token, err = jwt_service.GenerateJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка сервера. Попробуйте позже"},
			})
			return
		}

		c.SetCookie("token", token, 30*24*3600, "/", "localhost", false, true)

		c.JSON(http.StatusOK, gin.H{})

		go func() {
			if err := services.SendEmail(db, email); err != nil {
				log.Println("не удалось отправить письмо :( :", err)
			}
		}()
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

		var token string
		token, err = jwt_service.GenerateJWT(userID)

		c.SetCookie("token", token, 30*24*3600, "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{})
	})

	// Получение JWT-токена из Cookies для React'а
	r.GET("/api/me", func(c *gin.Context) {
		userID, err := services.GetUserIdFromJWT(c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка токена. Попробуйте перезайти в аккаунт"},
			})
		}

		username, email, isVerified, createdAt, err := services.DataAboutUserFromJWT(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка базы данных. Попробуйте позже"},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":     userID,
			"username":    username,
			"email":       email,
			"is_verified": isVerified,
			"created_at":  createdAt,
		})
	})

	r.GET("/verify/:code", func(c *gin.Context) {
		code := c.Param("code")

		err := services.VerifyEmail(db, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. Попробуйте позже",
			})
			return
		}

		userID, err := services.GetUserIDByCode(db, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. Попробуйте позже",
			})
			return
		}

		var token string
		token, err = jwt_service.GenerateJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка выдачи токена",
			})
		}

		c.SetCookie("token", token, 30*24*3600, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "http://localhost:5173/?verified=true")
	})

	r.POST("/api/resend-email", func(c *gin.Context) {
		var req ResendEmail
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		_, err = services.GetUserIdFromJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка токена. Попробуйте перезайти в аккаунт",
			})
			return
		}

		couldResend, time, err := services.CouldResendEmail(db, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. Попробуйте позже",
			})
			return
		}

		if couldResend == false {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Слишком частые запросы на отправку письма, подождите",
				"time_wait": int(time.Seconds()),
			})
			return
		}

		err = services.SendEmail(db, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка отправки письма. \n Попробуйте позже",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message_success": "Письмо успешно отправлено",
		})
	})

	r.POST("/api/reset-password/send-code", func(c *gin.Context) {
		var req ResendEmail
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		exist, err := services.IsEmailInDatabase(db, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		if exist == false {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Пользователь не найден",
			})
			return
		}

		couldResend, time, err := services.CouldResendEmail(db, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		seconds := int(time.Seconds())
		if couldResend == false {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf(
					"Слишком частые запросы на отправку письма, подождите %d %s",
					seconds,
					services.DeclinationWord(seconds, "секунда", "секунды", "секунд"),
				),
			})
			return
		}

		code := services.GenerateVerificationCode()
		err = services.PasswordResetEmail(req.Email, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка отправки письма. \n Попробуйте позже",
			})
			return
		}

		err = services.AddPasswordResetCodeToDatabase(db, req.Email, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			println(err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{})

	})

	r.POST("/api/reset-password/check-code", func(c *gin.Context) {
		var req CheckCode
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		exist, err := services.IsEmailInDatabase(db, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		if exist == false {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Пользователь не найден",
			})
			return
		}

		exist, err = services.CheckResetPasswordCode(db, req.Email, req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		if exist == false {
			attempts, err := services.DecreaseResetPasswordAttempts(db, req.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ошибка базы данных. \n Попробуйте позже",
				})
				return
			}

			if attempts == 0 {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":    "Попытки ввода закончились. Запросите новый код",
					"attempts": 0,
				})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Код введен неверно. \n Попыток осталось: ",
				"attempts": attempts,
			})
			return
		}

		isFresh, err := services.IsCodeFresh(db, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		if isFresh == false {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Код устарел, запросите новый",
			})
			return
		}

		resetToken, err := jwt_service.GenerateResetPasswordJWT(req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка сервера",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"reset_token": resetToken,
		})

	})

	r.POST("/api/reset-password", func(c *gin.Context) {
		var req ResetPassword
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		newPassword := strings.TrimSpace(req.NewPassword)
		confirmPassword := strings.TrimSpace(req.ConfirmPassword)

		if newPassword != confirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Пароли не совпадают",
			})
			return
		}

		var errs []string
		if len(newPassword) < 8 {
			errs = append(errs, "Минимальная длина пароля 8 символов")
		}
		if len(newPassword) > 72 {
			errs = append(errs, "Максимальная длина пароля 72 символа")
		}
		if !services.IsPasswordStrong(newPassword) {
			errs = append(errs, "Пароль должен содержать заглавную букву, строчную букву и цифру")
		}
		if !services.IsValidPassword(newPassword) {
			errs = append(errs, "Пароль содержит недопустимые символы")
		}

		if len(errs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}

		token, err := jwt.Parse(req.ResetToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Токен недействителен или истёк",
			})
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		if claims["purpose"] != "reset_password" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный тип токена",
			})
			return
		}

		email := claims["email"].(string)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка сервера. Попробуйте позже"},
			})
			return
		}

		err = services.ResetPassword(db, email, string(hashedPassword))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	r.GET("/api/qr/:code", func(c *gin.Context) {
		code := c.Param("code")

		_, err := services.ReturnOriginalLink(db, code)

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

		png, err := qrcode.Encode("http://localhost:8080/"+code, qrcode.Medium, 256)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка создания QR-кода",
			})
			return
		}

		c.Data(http.StatusOK, "image/png", png)
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("token", "", -1, "/", "localhost", false, true)

		c.JSON(http.StatusOK, gin.H{
			"message": "Вы вышли из аккаунта",
		})
	})

	r.DELETE("/delete-link/:id", func(c *gin.Context) {
		linkID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неверный ID ссылки",
			})
			return
		}

		userID, err := services.GetUserIdFromJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Ошибка токена. \n Попробуйте перезайти в аккаунт",
			})
			return
		}

		err = services.DeleteLink(db, linkID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка удаления ссылки",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	r.PATCH("/my-links/add-tag", func(c *gin.Context) {
		var req AddTag
		err := c.ShouldBindJSON(&req)
		fmt.Println(err)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		userID, err := services.GetUserIdFromJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Ошибка токена. Попробуйте перезайти в аккаунт",
			})
			return
		}

		couldAdd, err := services.CouldAddTag(db, userID, req.LinkID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		if !couldAdd {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "У вас нет прав изменять тэг",
			})
			return
		}

		tag := strings.TrimSpace(req.Tag)

		if utf8.RuneCountInString(tag) > 25 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Максимальная длина 25 символов",
			})
			return
		}

		err = services.AddTag(db, req.LinkID, tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			panic(err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})
}
