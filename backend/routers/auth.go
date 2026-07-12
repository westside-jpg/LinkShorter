package routers

import (
	jwt_service "LinkShorter/jwt"
	"LinkShorter/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

/*
Файл auth.go содержит роуты для регистрации, авторизации
и управления аккаунтом. Регистрация, вход, получение данных
о текущем пользователе, выход, верификация почты и повторная
отправка письма подтверждения
*/

func RegisterAuthRoutes(r *gin.Engine, db *pgxpool.Pool) {

	/*
		Регистрация нового пользователя. При успехе сразу выдаёт JWT
		в httpOnly cookie (логинит юзера сразу после регистрации,
		не дожидаясь верификации почты) и асинхронно отправляет
		письмо подтверждения через горутину
	*/
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

	// Вход по username или email, выдаёт JWT в httpOnly cookie
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
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{"Ошибка сервера. Попробуйте позже"},
			})
			return
		}

		c.SetCookie("token", token, 30*24*3600, "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{})
	})

	/*
		Возвращает данные текущего юзера (для восстановления
		сессии на фронте при перезагрузке страницы), userID
		извлекается из JWT в cookie
	*/
	r.GET("/api/me", func(c *gin.Context) {
		userID, err := services.GetUserIdFromJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Ошибка токена. Попробуйте перезайти в аккаунт"},
			})
			return
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

	/*
		Подтверждение почты по коду из письма. При успехе выдаёт
		новый JWT и редиректит на фронт с ?verified=true
		для отображения сообщения об успехе
	*/
	r.GET("/api/verify/:code", func(c *gin.Context) {
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
			return
		}

		c.SetCookie("token", token, 30*24*3600, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "http://localhost:5173/?verified=true")
	})

	/*
		Повторная отправка письма верификации почты.
		Ограничена rate limit в 60 секунд между отправками
	*/
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

	r.GET("/api/logout", func(c *gin.Context) {
		c.SetCookie("token", "", -1, "/", "localhost", false, true)

		c.JSON(http.StatusOK, gin.H{
			"message": "Вы вышли из аккаунта",
		})
	})
}
