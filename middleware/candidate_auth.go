package middleware

import (
	"net/http"
	"strings"

	"golang-backend/config"
	"golang-backend/entity"
	"golang-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CandidateAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, "Unauthorized", http.StatusUnauthorized, "Invalid authorization format")
			c.Abort()
			return
		}

		token, err := utils.ValidateToken(parts[1])
		if err != nil || !token.Valid {
			utils.ErrorResponse(c, "Unauthorized", http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["principal_type"] != "candidate" {
			utils.ErrorResponse(c, "Unauthorized", http.StatusUnauthorized, "Candidate token required")
			c.Abort()
			return
		}
		candidateID, ok := claims["candidate_id"].(string)
		if !ok || candidateID == "" {
			utils.ErrorResponse(c, "Unauthorized", http.StatusUnauthorized, "Invalid candidate identity")
			c.Abort()
			return
		}

		var candidate entity.Candidate
		if err := config.DB.Where("id = ? AND status = ?", candidateID, "active").First(&candidate).Error; err != nil {
			utils.ErrorResponse(c, "Unauthorized", http.StatusUnauthorized, "Candidate not found")
			c.Abort()
			return
		}
		candidate.PasswordHash = ""
		c.Set("candidate_id", candidate.ID)
		c.Set("currentCandidate", &candidate)
		c.Next()
	}
}
