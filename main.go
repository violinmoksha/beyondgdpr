// simple server with AES encrypt and decrypt endpoints
package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/gtank/cryptopasta"
)

func main() {
	debug := getEnv("DEBUG", "false")
	port := getEnv("PORT", "8080")
	basePath := getEnv("BASE_PATH", "")
	cfgFile := getEnv("CONFIG", "")

	// additional config for eventual k8s
	cfg := make(map[interface{}]interface{})
	if cfgFile != "" {
		ymlData, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal([]byte(ymlData), &cfg)
		if err != nil {
			panic(err)
		}
	}

	gin.SetMode(gin.ReleaseMode)

	if debug == "true" {
		gin.SetMode(gin.DebugMode)
	}

	logger, e := zap.NewProduction()
	if e != nil {
		panic(e.Error())
	}

	if debug == "true" {
		logger, _ = zap.NewDevelopment()
	}

	// router
	r := gin.New()
	rg := r.Group(basePath)

	// logger middleware
	rg.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	getUserKey := func(p string) *[32]byte {
		key := [32]byte{}
		_, err := io.ReadFull(bytes.NewReader([]byte(p)), key[:])
		if err != nil {
			panic(err)
		}

		return &key
	}

	// encryptPlaintext
	encryptPlaintext := func(c *gin.Context) {
		userkey := getUserKey(c.Query("userkey"))

		ciphertext, err := cryptopasta.Encrypt([]byte(c.Query("plaintext")), userkey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{"ciphertext": base64.URLEncoding.EncodeToString(ciphertext)})
		return
	}

	// decryptCiphertext
	decryptCiphertext := func(c *gin.Context) {
		userkey := getUserKey(c.Query("userkey"))

		decodedCiphertext, err := base64.URLEncoding.DecodeString(c.Query("ciphertext"))
		if err != nil {
			panic(err)
		}
		plaintext, err := cryptopasta.Decrypt(decodedCiphertext, userkey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{"plaintext": string(plaintext)})
		return
	}

	// for a k8s ingress
	aliveHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "alive"})
		return
	}

	// routes
	rg.POST("/encryptPlaintext", encryptPlaintext)
	rg.POST("/decryptCiphertext", decryptCiphertext)

	// for k8s livenessProbe
	rg.GET("/alive", aliveHandler)

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, false)
	})

	r.Run(":" + port)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}
