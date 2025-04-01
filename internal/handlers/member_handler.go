package handlers

import (
	"database/sql"
	"log"
	"membership/internal/db"
	"membership/internal/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// MemberHandler는 멤버 API 엔드포인트를 처리.
type MemberHandler struct {
	DB *db.MemberDB
}

// CreateMember는 새 멤버를 생성.
func (h *MemberHandler) CreateMember(c *gin.Context) {
	var member models.Member
	if err := c.BindJSON(&member); err != nil {
		log.Printf("Failed to bind request body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.InsertMember(member); err != nil {
		log.Printf("Failed to insert member: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create member"})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// GetMember는 ID로 멤버를 가져옴.
func (h *MemberHandler) GetMember(c *gin.Context) {
	id := c.Param("id")
	member, err := h.DB.GetMember(id)
	if err != nil {
		log.Printf("Failed to get member: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get member"})
		return
	}

	if member == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Member not found"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// UpdateMember는 멤버를 업데이트.
func (h *MemberHandler) UpdateMember(c *gin.Context) {
	id := c.Param("id")

	var member models.Member
	if err := c.BindJSON(&member); err != nil {
		log.Printf("Failed to bind request body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member.ID = id // URL 파라미터에서 ID 설정

	if err := h.DB.UpdateMember(member); err != nil {
		log.Printf("Failed to update member: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// DeleteMember는 멤버를 삭제.
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	id := c.Param("id")
	err := h.DB.DeleteMember(id)
	if err != nil {
		log.Printf("Failed to delete member: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member deleted"})
}

// SearchMembers는 이름으로 멤버를 검색.
func (h *MemberHandler) SearchMembers(c *gin.Context) {
	name := c.Query("name")
	members, err := h.DB.SearchMembers(name)
	if err != nil {
		log.Printf("Failed to search members: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search members"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// Login은 멤버의 로그인 요청을 처리.
func (h *MemberHandler) Login(c *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&loginRequest); err != nil {
		log.Printf("Failed to bind request body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 이메일로 멤버 조회.
	member, err := h.findMemberByEmail(loginRequest.Email)
	if err != nil {
		log.Printf("Failed to get member by email: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get member"})
		return
	}

	if member == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	// JWT 토큰을 생성.
	token, err := h.generateJWTToken(member.ID)
	if err != nil {
		log.Printf("Failed to generate JWT token: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// findMemberByEmail은 이메일 주소로 멤버 조회.
func (h *MemberHandler) findMemberByEmail(email string) (*models.Member, error) {
	var member models.Member
	query := "SELECT id, email, name, password, role, created_at FROM members WHERE email = ?"
	err := h.DB.Db.QueryRow(query, email).Scan(&member.ID, &member.Email, &member.Name, &member.Password, &member.Role, &member.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// generateJWTToken은 멤버 ID를 기반으로 JWT 토큰을 생성.
func (h *MemberHandler) generateJWTToken(memberID string) (string, error) {
	// 토큰 만료 시간 설정 (예: 1시간)
	expirationTime := time.Now().Add(1 * time.Hour)

	// JWT 클레임 생성
	claims := jwt.MapClaims{
		"member_id": memberID,
		"exp":       expirationTime.Unix(),
	}

	// JWT 토큰 생성
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// JWT 서명 키 가져오기 (환경 변수에서)
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		jwtKey = []byte("secret")
	}

	// JWT 토큰 서명
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
