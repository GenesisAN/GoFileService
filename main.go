package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const mb = 1024 * 1024

var workPath string

func main() {

	if err := loadEnv(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	workPath = filepath.Clean(os.Getenv("WORK_PATH"))
	r := gin.Default()
	r.Use(IPAndAuthorizationMiddleware())
	r.GET(os.Getenv("DOWNLOAD_RELATIVE_PATH")+"/*path", handleDownload)
	r.POST(os.Getenv("UPLOAD_RELATIVE_PATH"), handleUpload)
	if os.Getenv("HTTPS") == "true" {
		err := r.RunTLS(os.Getenv("ADDRESS"), filepath.Clean(os.Getenv("HTTPS_CERT_FILE")), filepath.Clean(os.Getenv("HTTPS_KEY_FILE")))
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

func loadEnv() error {
	var err error
	err = godotenv.Load()
	err = LoadConfig(filepath.Clean(os.Getenv("AUTH_CONFIG_FILE")))
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

	// Use filepath.Join to combine paths safely
	filePath := filepath.Join(workPath, decodedPath)

	// Ensure the path is still within the expected directory to prevent path traversal attacks
	if !strings.HasPrefix(filePath, workPath) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "invalid path"})
		return
	}

	fileInfo, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"status": "file not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error getting file info"})
		return
	}

	c.File(filePath)

	logMessage := logTransferDetails(c, "Downloaded", fullPath, fileInfo.Size(), startTime)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": logMessage})
}

func handleUpload(c *gin.Context) {
	startTime := time.Now()

	file, to, err := getUploadDetails(c)
	if err != nil {
		return
	}

	destPath := filepath.Clean(filepath.Join(workPath, to, file.Filename))

	if !strings.HasPrefix(destPath, workPath) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid path or filename"})
		return
	}

	// Check if the file already exists at the destination
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		uploadedFileHash, _ := hashFile(file)
		existingFileHash, _ := hashFileAtPath(destPath)

		c.JSON(http.StatusConflict, gin.H{
			"message":          "file already exists",
			"existingFileHash": existingFileHash,
			"uploadedFileHash": uploadedFileHash,
		})
		return
	}

	if err = c.SaveUploadedFile(file, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	logMessage := logTransferDetails(c, "Uploaded", file.Filename, file.Size, startTime)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": logMessage})
}

func hashFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashFileAtPath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
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

func logTransferDetails(c *gin.Context, action, path string, size int64, startTime time.Time) string {
	duration := time.Since(startTime).Seconds()
	speed := float64(size) / (mb * duration)
	message := fmt.Sprintf("IP: %s | %s file: %s | Size: %d bytes | Speed: %.2f MB/s", c.ClientIP(), action, path, size, speed)
	log.Println(message)
	return message
}
