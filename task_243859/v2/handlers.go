
package main

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
)

func (app *Application) registerHandler(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := app.UserSvc.Create(&user); err != nil {
        switch err {
        case ErrDuplicateUsername:
            c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
        case ErrDuplicateEmail:
            c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        }
        return
    }

    user.Password = ""
    c.JSON(http.StatusCreated, user)
}

func (app *Application) loginHandler(c *gin.Context) {
    var credentials struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&credentials); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := app.UserSvc.Authenticate(credentials.Username, credentials.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    session := sessions.Default(c)
    session.Set("authenticated", true)
    session.Set("user_id", user.ID)
    if err := session.Save(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
        return
    }

    c.JSON(http.StatusOK, user)
}

func (app *Application) listUsersHandler(c *gin.Context) {
    users, err := app.UserSvc.List()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
        return
    }

    c.JSON(http.StatusOK, users)
}

func (app *Application) getUserHandler(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    user, err := app.UserSvc.GetByID(id)
    if err != nil {
        switch err {
        case ErrUserNotFound:
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
        }
        return
    }

    c.JSON(http.StatusOK, user)
}

func (app *Application) updateUserHandler(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user.ID = id
    if err := app.UserSvc.Update(&user); err != nil {
        switch err {
        case ErrUserNotFound:
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        case ErrDuplicateUsername:
            c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
        case ErrDuplicateEmail:
            c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        }
        return
    }

    c.JSON(http.StatusOK, user)
}

func (app *Application) deleteUserHandler(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    if err := app.UserSvc.Delete(id); err != nil {
        switch err {
        case ErrUserNotFound:
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        }
        return
    }

    c.Status(http.StatusOK)
}

func (app *Application) authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        if auth, _ := session.Get("authenticated").(bool); !auth {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
        c.Next()
    }
}
