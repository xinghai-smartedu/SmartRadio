package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// QQMusicAPI 封装QQ音乐相关API操作
type QQMusicAPI struct{}

// SongInfo 表示从QQ音乐获取的歌曲信息
type SongInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"` // 歌曲时长（秒）
	URL      string `json:"url"`      // 音频直链
}

// SearchSongs 搜索歌曲
func (q *QQMusicAPI) SearchSongs(keyword string, limit int) ([]SongInfo, error) {
	searchURL := fmt.Sprintf(
		"https://c.y.qq.com/soso/fcgi-bin/client_search_cp?ct=24&qqmusic_ver=12981080&remoteplace=txt.yqq.center&searchid=71671388493344141&t=0&aggr=1&cr=1&catZhida=1&lossless=0&flag_qc=0&p=1&n=%d&w=%s&g_tk=5381&loginUin=0&hostUin=0&format=json&inCharset=utf8&outCharset=utf-8&notice=0&platform=yqq.json&needNewCode=0",
		limit,
		url.QueryEscape(keyword),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search songs: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// 解析搜索结果
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	list, ok := data["song"].(map[string]interface{})["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no songs found")
	}

	var songs []SongInfo
	for _, item := range list {
		songData := item.(map[string]interface{})

		// 处理songmid，它可能是字符串或数字
		var songID string
		if mid, ok := songData["songmid"].(string); ok {
			songID = mid
		} else if midFloat, ok := songData["songmid"].(float64); ok {
			songID = fmt.Sprintf("%.0f", midFloat)
		} else {
			// 如果都没有找到，跳过这首歌
			continue
		}

		title := songData["songname"].(string)
		artistList := songData["singer"].([]interface{})
		artistNames := make([]string, len(artistList))
		for i, singer := range artistList {
			singerData := singer.(map[string]interface{})
			artistNames[i] = singerData["name"].(string)
		}
		artist := strings.Join(artistNames, ", ")

		album := ""
		if albumData, ok := songData["albumname"].(string); ok {
			album = albumData
		} else if albumFloat, ok := songData["albumname"].(float64); ok {
			album = fmt.Sprintf("%.0f", albumFloat)
		}

		duration := 0
		if durationData, ok := songData["interval"].(float64); ok {
			duration = int(durationData)
		} else if durationStr, ok := songData["interval"].(string); ok {
			// 尝试将字符串转换为整数
			fmt.Sscanf(durationStr, "%d", &duration)
		}

		song := SongInfo{
			ID:       songID,
			Title:    title,
			Artist:   artist,
			Album:    album,
			Duration: duration,
		}
		songs = append(songs, song)
	}

	return songs, nil
}

// GetSongURL 获取歌曲直链（注意：由于版权保护，直接获取可能不可行）
func (q *QQMusicAPI) GetSongURL(songmid string) (string, error) {
	getMusicURL := "https://zy.ricuo.com/"
	client := &http.Client{Timeout: 10 * time.Second}
	// 发送POST请求获取歌曲直链
	resp, err := client.Post(getMusicURL, "application/json", strings.NewReader(`{"input":"`+songmid+`", "filter":"id", "type":"qq", "page":1}`))
	if err != nil {
		// 返回码为无
		return "", fmt.Errorf("failed to get song URL: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查响应是否成功
	code, ok := result["code"].(float64)
	if !ok || int(code) != 200 {
		errorMsg, _ := result["error"].(string)
		return "", fmt.Errorf("API error: %s", errorMsg)
	}

	// 解析歌曲URL - 响应中的data是一个数组
	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return "", fmt.Errorf("invalid response format or empty data")
	}

	// 获取第一个结果
	firstResult, ok := data[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	url, ok := firstResult["url"].(string)
	if !ok {
		return "", fmt.Errorf("no URL found in response")
	}

	return url, nil
}

// GlobalQQMusicAPI 全局QQ音乐API实例
var GlobalQQMusicAPI = &QQMusicAPI{}
