package routers

import (
	jwt_service "LinkShorter/jwt"
	"LinkShorter/services"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

/*
Файл reset_password.go содержит роуты для полного цикла
сброса пароля в три шага. Отправка кода на почту, проверка
кода с учётом попыток и срока действия, и финальная смена
пароля по reset_token
*/

func RegisterResetPasswordRoutes(r *gin.Engine, db *pgxpool.Pool) {

	/*
		Отправка кода подтверждения на почту для сброса пароля.
		Ограничена rate limit в 60 секунд между отправками. Код
		сначала сохраняется в базу, потом отправляется письмом,
		чтобы не оказаться в ситуации с отправленным письмом
		и несохранённым кодом
	*/
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
		err = services.AddPasswordResetCodeToDatabase(db, req.Email, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		err = services.SendPasswordResetEmail(req.Email, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка отправки письма. \n Попробуйте позже",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	/*
		Проверка кода из письма. При неверном коде уменьшает
		счётчик оставшихся попыток (reset_password_attempts в БД),
		при исчерпании попыток требует запросить новый код.
		При верном и свежем коде выдаёт отдельный reset_token
		(JWT с email и коротким сроком жизни 15 минут), который
		затем используется в /api/reset-password для смены пароля
	*/
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

	/*
		Финальная смена пароля. Email берётся не из запроса,
		а из reset_token (claims["email"]), чтобы клиент не мог
		подменить, чей пароль меняется. Токен дополнительно
		проверяется на purpose="reset_password", чтобы обычный
		JWT авторизации нельзя было использовать вместо reset_token
	*/
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

}
