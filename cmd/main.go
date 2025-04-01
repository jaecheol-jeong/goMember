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

	"github.com/dgrijalva/jwt-go"
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

	// 미들웨어 설정: JWT 인증
	authorized := router.Group("/api")
	authorized.Use(AuthMiddleware())
	{
		// 로그인된 사용자만 접근 가능한 Route
		authorized.POST("/members", memberHandler.CreateMember)
		authorized.GET("/members/:id", memberHandler.GetMember)
		authorized.PUT("/members/:id", memberHandler.UpdateMember)
		authorized.DELETE("/members/:id", memberHandler.DeleteMember)
		authorized.GET("/members/search", memberHandler.SearchMembers)
	}

	// Login Route (인증 미들웨어 제외)
	router.POST("/login", memberHandler.Login)

	// Health Check 엔드포인트 (인증 미들웨어 제외)
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
		port = "8001" // 기본 포트
	}
	log.Printf("Server is running on port %s", port)
	if err := router.Run(":" + port); err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %s", err)
	}
}

// AuthMiddleware는 JWT 토큰을 검증하는 미들웨어입니다.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization 헤더에서 토큰을 추출합니다.
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header required"})
			return
		}

		// 토큰 검증
		token, err := ValidateJWTToken(tokenString)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			return
		}

		// 토큰에서 사용자 ID 추출 (필요한 경우)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to extract claims"})
			return
		}

		memberID, ok := claims["member_id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Invalid member ID"})
			return
		}

		// 사용자 ID를 컨텍스트에 저장 (필요한 경우)
		c.Set("member_id", memberID)

		c.Next()
	}
}

// ValidateJWTToken은 JWT 토큰을 검증합니다.
func ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		jwtKey = []byte("secret")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 서명 방법 검증
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
