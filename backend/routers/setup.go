package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Файл setup.go служит точкой входа для регистрации всех
роутов приложения. Собирает воедино роуты из auth.go,
links.go и reset_password.go, вызывается один раз из main.go
*/

func SetupRoutes(r *gin.Engine, db *pgxpool.Pool) {
	/*
		Порядок вызова важен из-за /:code в links.go. Это catch-all
		роут для одиночного сегмента пути, поэтому RegisterLinkRoutes
		должен вызываться последним, иначе /:code перехватит на себя
		остальные одиночные GET-пути. При добавлении нового такого
		пути в auth.go или reset_password.go, нужно регистрировать
		его до RegisterLinkRoutes, а не после
	*/
	RegisterAuthRoutes(r, db)
	RegisterResetPasswordRoutes(r, db)
	RegisterLinkRoutes(r, db)
}
