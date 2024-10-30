
package main

import (
    "database/sql"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    _ "github.com/go-sql-driver/mysql"
)

type Application struct {
    DB      *sql.DB
    Router  *gin.Engine
    UserSvc UserService
}

func NewApplication() (*Application, error) {
    db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/crud_db?parseTime=true")
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    router := gin.Default()
    store := cookie.NewStore([]byte("your-secret-key"))
    router.Use(sessions.Sessions("mysession", store))

    app := &Application{
        DB:      db,
        Router:  router,
        UserSvc: &SQLUserService{db: db},
    }

    app.setupRoutes()

    return app, nil
}

func (app *Application) setupRoutes() {
    app.Router.POST("/register", app.registerHandler)
    app.Router.POST("/login", app.loginHandler)

    protected := app.Router.Group("/")
    protected.Use(app.authMiddleware())
    {
        protected.GET("/users", app.listUsersHandler)
        protected.GET("/users/:id", app.getUserHandler)
        protected.PUT("/users/:id", app.updateUserHandler)
        protected.DELETE("/users/:id", app.deleteUserHandler)
    }
}
