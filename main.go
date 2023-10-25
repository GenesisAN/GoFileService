package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

const mb = 1024 * 1024

var workPath = os.Getenv("WORK_PATH")
var AuthIP map[string]bool
var AuthorizationHeader string

func main() {

	if err := loadEnv(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := gin.Default()
	r.Use(IPAndAuthorizationMiddleware())
	r.GET(os.Getenv("DOWNLOAD_RELATIVE_PATH")+"/*path", handleDownload)
	r.POST(os.Getenv("UPLOAD_RELATIVE_PATH"), handleUpload)
	if os.Getenv("HTTPS") == "true" {
		err := r.RunTLS(os.Getenv("ADDRESS"), os.Getenv("HTTPS_CERT_FILE"), os.Getenv("HTTPS_KEY_FILE"))
		if err != nil {
			panic(err)
		}
	} else {
		err := r.Run(os.Getenv("ADDRESS"))
		if err != nil {
			panic(err)
		}
	}

}

var authconfig *AutoConfig

func loadEnv() error {
	var err error
	err = godotenv.Load()
	LoadConfig(os.Getenv("AUTH_CONFIG_FILE"))
	return err
}

func handleDownload(c *gin.Context) {
	startTime := time.Now()

	fullPath := c.Param("path")
	decodedPath, err := url.QueryUnescape(fullPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "invalid filename"})
		return
	}

	// 使用decodedPath替代fullPath来访问文件
	filePath := os.Getenv("WORK_PATH") + decodedPath

	if !fileExists(filePath) {
		c.JSON(http.StatusNotFound, gin.H{"status": "file not found"})
		return
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error getting file info"})
		return
	}

	c.File(filePath)

	logMessage := logTransferDetails(c, "Downloaded", fullPath, fileInfo.Size(), startTime)
	c.String(http.StatusOK, logMessage)
}

func handleUpload(c *gin.Context) {
	startTime := time.Now()

	file, to, err := getUploadDetails(c)
	if err != nil {
		return
	}

	destPath := workPath + to + file.Filename

	if err = c.SaveUploadedFile(file, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	logMessage := logTransferDetails(c, "Uploaded", file.Filename, file.Size, startTime)
	c.String(http.StatusOK, "File uploaded successfully!\n"+logMessage)
}

func getUploadDetails(c *gin.Context) (*multipart.FileHeader, string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return nil, "", err
	}

	to, has := c.GetPostForm("to")
	if !has {
		c.JSON(http.StatusBadRequest, gin.H{"message": "to is not found"})
		return nil, "", err
	}

	return file, to, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func logTransferDetails(c *gin.Context, action, path string, size int64, startTime time.Time) string {
	duration := time.Since(startTime).Seconds()
	speed := float64(size) / (mb * duration)
	message := fmt.Sprintf("IP: %s | %s file: %s | Size: %d bytes | Speed: %.2f MB/s", c.ClientIP(), action, path, size, speed)
	log.Println(message)
	return message
}
func LoadConfig(filename string) error {
	var config AutoConfig

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	// 解析YAML内容到结构体中
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	AuthIP = make(map[string]bool)
	for _, ip := range config.AuthorizedIPs {
		AuthIP[ip] = true
	}
	AuthorizationHeader = config.AuthorizationHeader
	return nil
}

type AutoConfig struct {
	AuthorizedIPs       []string `yaml:"AuthorizedIPs"`
	AuthorizationHeader string   `yaml:"AuthorizationHeader"`
}

func IPAndAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !AuthIP[clientIP] {
			// 如果IP不在授权列表中，检查授权码
			authorizationCode := c.GetHeader("Authorization")
			if authorizationCode != AuthorizationHeader {
				c.JSON(http.StatusForbidden, gin.H{"status": "unauthorized"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
