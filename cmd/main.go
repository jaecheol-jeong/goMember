package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"membership/internal/db"
	"membership/internal/handlers"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type DBConfig struct {
	USERNAME string `json:"USERNAME"`
	PASSWORD string `json:"PASSWORD"`
	PORT     string `json:"PORT"`
	HOST     string `json:"HOST"`
	DB_NAME  string `json:"DB_NAME"`
	KIND     string `json:"KIND"`
}

type AppConfig struct {
	APP_NAME string   `json:"APP_NAME"`
	APP_PORT string   `json:"APP_PORT"`
	DB       DBConfig `json:"DB"`
}

func main() {
	// config.json 파일 읽기
	jsonFile, err := os.Open("config/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var config AppConfig

	json.Unmarshal(byteValue, &config)

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		config.DB.USERNAME, config.DB.PASSWORD, config.DB.HOST, config.DB.PORT, config.DB.DB_NAME)

	memberDB, err := db.NewMemberDB(dataSourceName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err)
		os.Exit(1)
	}

	// 연결 풀 설정
	memberDB.Db.SetMaxIdleConns(10)           // 최대 유휴 연결 수
	memberDB.Db.SetMaxOpenConns(100)          // 최대 열린 연결 수
	memberDB.Db.SetConnMaxLifetime(time.Hour) // 연결 최대 생존 시간

	defer memberDB.Close()

	router := gin.Default()

	// 핸들러 초기화
	memberHandler := &handlers.MemberHandler{DB: memberDB}

	// Routes
	router.POST("/members", memberHandler.CreateMember)
	router.GET("/members/:id", memberHandler.GetMember)
	router.PUT("/members/:id", memberHandler.UpdateMember)
	router.DELETE("/members/:id", memberHandler.DeleteMember)
	router.GET("/members/search", memberHandler.SearchMembers)

	// Health Check 엔드포인트
	router.GET("/health", func(c *gin.Context) {
		err := memberDB.Db.Ping()
		if err != nil {
			log.Printf("Database ping failed: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// 서버 시작
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // 기본 포트
	}
	log.Printf("Server is running on port %s", port)
	if err := router.Run(":" + port); err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %s", err)
	}
}
