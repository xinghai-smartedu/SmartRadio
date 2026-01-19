package main

import (
	"net/http"
	"strconv"
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

		// QQ音乐API路由
		qqmusic := api.Group("/qqmusic")
		{
			qqmusic.GET("/search", searchSongsFromQQ)
			qqmusic.GET("/play/:songmid", getSongURL)
			qqmusic.POST("/add/:songmid", addSongFromQQ)
		}
	}

	// 前端页面路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.html", nil)
	})

	r.Run(":8080")
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

	// 如果提供了songmid但没有提供URL，则尝试获取URL
	if song.URL == "" {
		// 这里可以扩展逻辑来根据其他信息获取URL
		// 暂时保持空URL
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

// 从QQ音乐搜索歌曲
func searchSongsFromQQ(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keyword is required"})
		return
	}

	limitStr := c.Query("limit")
	limit := 10 // 默认返回10首
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 50 { // 限制最大数量
				limit = 50
			} else {
				limit = parsedLimit
			}
		}
	}

	songs, err := GlobalQQMusicAPI.SearchSongs(keyword, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"songs": songs})
}

// 通过QQ音乐ID直接添加歌曲
func addSongFromQQ(c *gin.Context) {
	songmid := c.Param("songmid")

	// 获取歌曲详细信息
	songs, err := GlobalQQMusicAPI.SearchSongs(songmid, 1)
	if err != nil || len(songs) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get song details"})
		return
	}

	songDetail := songs[0]

	// 获取歌曲URL
	url, err := GlobalQQMusicAPI.GetSongURL(songmid)
	if err != nil {
		// 即使无法获取URL，也可以添加歌曲，只是URL字段为空
		url = ""
	}

	// 创建新的点歌请求，包含URL
	newSong := SongRequest{
		Title:     songDetail.Title,
		Artist:    songDetail.Artist,
		Requester: "QQMusic",
		Status:    "pending",
		URL:       url,
	}

	createdSong := store.Create(newSong)
	c.JSON(http.StatusOK, createdSong)
}

// 获取歌曲播放链接
func getSongURL(c *gin.Context) {
	songmid := c.Param("songmid")

	url, err := GlobalQQMusicAPI.GetSongURL(songmid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
