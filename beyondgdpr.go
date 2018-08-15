// Package beyondgdpr is a simple server with AES encrypt and decrypt endpoints
package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/gtank/cryptopasta"
)

var wg sync.WaitGroup

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

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

	getUserKey := func(p string) (*[32]byte, error) {
		key := [32]byte{}
		_, readErr := io.ReadFull(bytes.NewReader([]byte(p)), key[:])
		if readErr != nil {
			return nil, readErr
		}

		return &key, nil
	}

	// encryptPlaintext
	encryptPlaintext := func(threadC *gin.Context) {
		type PlaintextForEncrypt struct {
			Plaintext string `form:"plaintext" json:"plaintext" binding:"required"`
			Userkey   string `form:"userkey" json:"userkey" binding:"required"`
		}

		var plaintextJSON PlaintextForEncrypt
		if err := threadC.ShouldBindJSON(&plaintextJSON); err == nil {
			if utf8.RuneCountInString(plaintextJSON.Userkey) != 44 {
				threadC.JSON(http.StatusBadRequest, gin.H{"Error": "userkey should be [32]byte as [44]rune"})
				wg.Done()
			} else {
				userkey, keyErr := getUserKey(plaintextJSON.Userkey)
				if keyErr != nil {
					threadC.JSON(http.StatusInternalServerError, keyErr.Error())
					wg.Done()
				}

				ciphertext, cryptErr := cryptopasta.Encrypt([]byte(plaintextJSON.Plaintext), userkey)
				if cryptErr != nil {
					threadC.JSON(http.StatusInternalServerError, cryptErr.Error())
					wg.Done()
				}

				threadC.JSON(http.StatusOK, gin.H{"ciphertext": base64.URLEncoding.EncodeToString(ciphertext)})
				wg.Done()
			}
		} else {
			threadC.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			wg.Done()
		}
	}

	// decryptCiphertext
	decryptCiphertext := func(threadC *gin.Context) {
		type CiphertextForEncrypt struct {
			Ciphertext string `form:"ciphertext" json:"ciphertext" binding:"required"`
			Userkey    string `form:"userkey" json:"userkey" binding:"required"`
		}

		var cipherJSON CiphertextForEncrypt
		if err := threadC.ShouldBindJSON(&cipherJSON); err == nil {
			if utf8.RuneCountInString(cipherJSON.Userkey) != 44 {
				threadC.JSON(http.StatusBadRequest, gin.H{"Error": "userkey should be [32]byte as [44]rune"})
				wg.Done()
			} else {
				userkey, keyErr := getUserKey(cipherJSON.Userkey)
				if keyErr != nil {
					threadC.JSON(http.StatusInternalServerError, keyErr.Error())
				}
				decodedCiphertext, decodeCipherErr := base64.URLEncoding.DecodeString(cipherJSON.Ciphertext)
				if decodeCipherErr != nil {
					logger.Info("Some issue decoding ciphertext", zap.Error(decodeCipherErr))
				}
				plaintext, decryptErr := cryptopasta.Decrypt(decodedCiphertext, userkey)
				if decryptErr != nil {
					threadC.JSON(http.StatusInternalServerError, decryptErr.Error())
					wg.Done()
				}

				threadC.JSON(http.StatusOK, gin.H{"plaintext": string(plaintext)})
				wg.Done()
			}
		} else {
			threadC.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			wg.Done()
		}
	}

	// for a k8s ingress
	aliveHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "alive"})
		return
	}

	// routes
	rg.POST("/encryptPlaintext", func(c *gin.Context) {
		wg.Add(1)
		go encryptPlaintext(c)
		wg.Wait()
	})
	rg.POST("/decryptCiphertext", func(c *gin.Context) {
		wg.Add(1)
		go decryptCiphertext(c)
		wg.Wait()
	})

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
