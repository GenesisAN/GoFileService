package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
)

type AutoConfig struct {
	AuthorizedIPs       []string `yaml:"AuthorizedIPs"`
	AuthorizationHeader string   `yaml:"AuthorizationHeader"`
}

var AuthIP map[string]bool
var AuthorizationHeader string

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
