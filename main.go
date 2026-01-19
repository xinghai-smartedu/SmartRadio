package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// SongRequest 表示点歌请求
type SongRequest struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Artist    string    `json:"artist" binding:"required"`
	Requester string    `json:"requester" binding:"required"`
	Status    string    `json:"status"` // pending, playing, played
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url,omitempty"` // 音乐直链（可选）
}

var store *JSONStore

func main() {
	// 初始化JSON存储
	var err error
	store, err = NewJSONStore("data.json")
	if err != nil {
		panic("failed to initialize JSON store: " + err.Error())
	}

	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// API路由组
	api := r.Group("/api")
	{
		api.GET("/songs", getSongs)
		api.POST("/songs", addSong)
		api.PUT("/songs/:id/status", updateSongStatus)
		api.DELETE("/songs/:id", deleteSong)

	}

	// 前端页面路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.html", nil)
	})

	// 创建一个通道来接收系统中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 在一个goroutine中启动HTTP服务器
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	log.Println("服务器启动在 :8080 端口，等待中断信号...")

	// 等待中断信号
	sig := <-sigChan
	log.Printf("收到终止信号: %v", sig)
	log.Println("正在保存数据...")

	// 尝试保存临时文件
	saveTempJSONFiles()

	log.Println("数据保存完成")
	os.Exit(0)
}

// 保存临时JSON文件到正式文件
func saveTempJSONFiles() {
	// 检查是否存在临时文件
	tempFile := "data.json.tmp"
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		log.Printf("临时文件 %s 不存在，跳过保存", tempFile)
		return
	}

	// 读取临时文件内容
	tempData, err := os.ReadFile(tempFile)
	if err != nil {
		log.Printf("读取临时文件失败: %v", err)
		// 如果临时文件读取失败，尝试读取原文件
		originalData, readErr := os.ReadFile("data.json")
		if readErr != nil {
			log.Printf("原文件也读取失败: %v", readErr)
			return
		}
		tempData = originalData
	}

	// 将内容写入正式文件
	if err := os.WriteFile("data.json", tempData, 0644); err != nil {
		log.Printf("写入数据文件失败: %v", err)
		return
	}

	log.Println("数据已成功保存到 data.json")
}

// 获取所有点歌请求
func getSongs(c *gin.Context) {
	status := c.Query("status")
	if status != "" {
		songs := store.GetByStatus(status)
		c.JSON(http.StatusOK, songs)
		return
	}
	songs := store.GetAll()
	c.JSON(http.StatusOK, songs)
}

// 添加新的点歌请求
func addSong(c *gin.Context) {
	var song SongRequest
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	song = store.Create(song)
	c.JSON(http.StatusOK, song)
}

// 更新歌曲状态
func updateSongStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,eq=pending|eq=playing|eq=played"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedSong, found := store.UpdateStatus(uint(id), req.Status)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	c.JSON(http.StatusOK, updatedSong)
}

// 删除点歌请求
func deleteSong(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	found := store.Delete(uint(id))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song deleted successfully"})
}
