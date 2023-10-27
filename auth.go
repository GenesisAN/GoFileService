package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"os"
	"strings"
)

type AutoConfig struct {
	AuthorizedIPs       []string `yaml:"AuthorizedIPs"`
	AuthorizationHeader string   `yaml:"AuthorizationHeader"`
}

var AuthIPRanges []*net.IPNet
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
	AuthIPRanges = make([]*net.IPNet, len(config.AuthorizedIPs))
	for i, ipOrCIDR := range config.AuthorizedIPs {
		if !strings.Contains(ipOrCIDR, "/") {
			// 如果它是一个单独的IP地址，自动添加/32
			ipOrCIDR = ipOrCIDR + "/32"
		}
		_, ipNet, err := net.ParseCIDR(ipOrCIDR)
		if err != nil {
			return err
		}
		AuthIPRanges[i] = ipNet
	}
	AuthorizationHeader = config.AuthorizationHeader
	return nil
}

func IPAndAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := net.ParseIP(c.ClientIP())
		authorized := false
		for _, ipRange := range AuthIPRanges {
			if ipRange.Contains(clientIP) {
				authorized = true
				break
			}
		}
		if !authorized {
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
