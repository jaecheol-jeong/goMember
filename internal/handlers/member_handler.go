package handlers

import (
	"log"
	"membership/internal/db"
	"membership/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MemberHandler는 멤버 API 엔드포인트를 처리합니다.
type MemberHandler struct {
	DB *db.MemberDB
}

// CreateMember는 새 멤버를 생성합니다.
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

// GetMember는 ID로 멤버를 가져옵니다.
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

// UpdateMember는 멤버를 업데이트합니다.
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

// DeleteMember는 멤버를 삭제합니다.
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

// SearchMembers는 이름으로 멤버를 검색합니다.
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
