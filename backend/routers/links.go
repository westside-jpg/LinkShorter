package routers

import (
	"LinkShorter/models"
	"LinkShorter/services"
	"errors"
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
	"github.com/skip2/go-qrcode"
)

/*
Файл links.go содержит роуты для работы со ссылками.
Создание коротких и кастомных ссылок, получение списка
ссылок пользователя с сортировкой и фильтрацией, генерация
QR-кода, удаление ссылки, добавление тега и редирект
по короткому коду
*/

func RegisterLinkRoutes(r *gin.Engine, db *pgxpool.Pool) {

	/*
		Сокращение ссылки. Работает и для анонимов (без cookie,
		ссылка попадает в общий анонимный пул с user_id=0), и для
		авторизованных юзеров (личный пул, JWT парсится вручную
		из cookie прямо здесь, а не через GetUserIdFromJWT)
	*/
	r.POST("/api/create-link", func(c *gin.Context) {
		var req CreateLinkRequest
		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		url, err := services.ValidateURL(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Создание ссылки для незарегистрированного пользователя
		tokenString, err := c.Cookie("token")
		if err != nil {
			shortLink, err := services.GenerateLink(db, url, 0)
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

		shortLink, err := services.GenerateLink(db, url, userID)
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

	/*
		Создание кастомной короткой ссылки с именем, заданным
		юзером. Доступно, только если человек авторизован
	*/
	r.POST("/api/create-link/custom", func(c *gin.Context) {
		var req CreateCustomLinkRequest
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неправильный запрос",
			})
			return
		}

		url, err := services.ValidateURL(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		custom, err := services.ValidateCustomLink(req.Custom)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
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

		shortLink, exist, err := services.CreateCustomLink(db, userID, url, custom)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		if exist {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Эта ссылка уже занята",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"short-link": shortLink,
		})

	})

	/*
		Список ссылок текущего юзера с поддержкой поиска, сортировки
		и фильтрации через query-параметры (sort, order, period,
		views, tags, search)
	*/
	r.GET("/api/my-links", func(c *gin.Context) {
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
		filterViews := c.DefaultQuery("views", "0")
		filterTags := c.DefaultQuery("tags", "all")

		links, err := services.UserLinks(db, userID, search, sortBy, orderBy, filterPeriod, filterViews, filterTags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Ошибка базы данных",
			})
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

	// Генерация QR-кода для короткой ссылки в формате PNG
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

	// Удаление ссылки
	r.DELETE("/api/delete-link/:id", func(c *gin.Context) {
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

		couldDelete, err := services.IsLinkOwner(db, userID, linkID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных. \n Попробуйте позже",
			})
			return
		}

		if !couldDelete {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "У вас нет прав удалять ссылку",
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

	// Добавление или обновление тега на ссылке
	r.PATCH("/api/my-links/add-tag", func(c *gin.Context) {
		var req AddTag
		err := c.ShouldBindJSON(&req)
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

		couldAdd, err := services.IsLinkOwner(db, userID, req.LinkID)
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
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	/*
		Редирект по короткой ссылке на оригинальный URL.
		Обрабатывает любой одиночный путь, поэтому регистрируется
		последним, чтобы не перехватывать более специфичные роуты выше
		(подробности про порядок регистрации в setup.go)
	*/
	r.GET("/:code", func(c *gin.Context) {
		code := c.Param("code")
		originalURL, err := services.ReturnOriginalLink(db, code)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.Redirect(http.StatusFound, "http://localhost:5173/link-not-found")
				return
			}
			c.Redirect(http.StatusFound, "http://localhost:5173/server-error")
			return
		}

		err = services.IncreaseLinkViews(db, code)
		if err != nil {
			log.Println("Не удалось увеличить счетчик просмотров:", err)
		}

		c.Redirect(http.StatusFound, originalURL)
	})
}
