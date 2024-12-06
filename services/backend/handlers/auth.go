package handlers

import (
	"net/http"
	"strconv"

	"go-turbo/pkg/auth"
	"go-turbo/pkg/database"
	"go-turbo/pkg/events"
	"go-turbo/pkg/models"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	db        *database.Database
	publisher *events.Publisher
}

func NewAuthHandler(db *database.Database, publisher *events.Publisher) *AuthHandler {
	return &AuthHandler{
		db:        db,
		publisher: publisher,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user, err := models.ValidateUserCredentials(c.Request.Context(), h.db.Pool, loginReq.Email, loginReq.Password)
	if err != nil {
		// Track failed login attempt
		h.publisher.TrackLogin(c.Request.Context(), 0, false, map[string]string{
			"email": loginReq.Email,
			"error": err.Error(),
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateJWT(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Track successful login
	h.publisher.TrackLogin(c.Request.Context(), uint(user.ID), true, map[string]string{
		"email": user.Email,
		"role":  user.Role,
	})

	// Log audit event
	h.publisher.LogUserAction(c.Request.Context(), uint(user.ID), models.ActionLogin, models.ResourceUser, strconv.FormatUint(uint64(user.ID), 10), map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	})

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if user.Role == "" {
		user.Role = "user" // Default role
	}

	if err := models.CreateUser(c.Request.Context(), h.db.Pool, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	// Track registration
	h.publisher.TrackRegistration(c.Request.Context(), uint(user.ID), map[string]string{
		"email": user.Email,
		"role":  user.Role,
	})

	// Log audit event
	h.publisher.LogUserAction(c.Request.Context(), uint(user.ID), models.ActionCreate, models.ResourceUser, strconv.FormatUint(uint64(user.ID), 10), map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	})

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) GetUsers(c *gin.Context) {
	users, err := models.GetAllUsers(c.Request.Context(), h.db.Pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}

	// Log audit event
	userID := c.GetUint("userID")
	h.publisher.LogUserAction(c.Request.Context(), userID, models.ActionRead, models.ResourceUser, "all", nil)

	c.JSON(http.StatusOK, users)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := models.GetUserByID(c.Request.Context(), h.db.Pool, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Log audit event
	h.publisher.LogUserAction(c.Request.Context(), userID, models.ActionRead, models.ResourceProfile, strconv.FormatUint(uint64(user.ID), 10), nil)

	user.Password = "" // Don't send password back
	c.JSON(http.StatusOK, user)
}
