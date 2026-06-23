package routers

import (
	"LinkShorter/services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateLinkRequest struct {
	URL string `json:"url"`
}

func SetupRoutes(r *gin.Engine, db *pgxpool.Pool) {

	r.POST("/create-link", func(c *gin.Context) {
		var req CreateLinkRequest
		err := c.ShouldBindJSON(&req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
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
}
