package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// JSONStore 定义JSON存储结构
type JSONStore struct {
	mu       sync.RWMutex
	data     []SongRequest
	filename string
	nextID   uint
}

// NewJSONStore 创建新的JSON存储实例
func NewJSONStore(filename string) (*JSONStore, error) {
	store := &JSONStore{filename: filename}
	err := store.loadFromFile()
	return store, err
}

// loadFromFile 从文件加载数据
func (j *JSONStore) loadFromFile() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	file, err := os.Open(j.filename)
	if err != nil {
		// 如果文件不存在，则初始化为空数据
		if os.IsNotExist(err) {
			j.data = []SongRequest{}
			j.nextID = 1
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&j.data)
	if err != nil {
		return err
	}

	// 计算下一个ID
	j.nextID = 1
	for _, item := range j.data {
		if item.ID >= j.nextID {
			j.nextID = item.ID + 1
		}
	}

	return nil
}

// saveToFile 将数据保存到文件
func (j *JSONStore) saveToFile() error {
	j.mu.RLock()
	defer j.mu.RUnlock()

	tempFilename := j.filename + ".tmp"
	file, err := os.Create(tempFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(j.data)
	if err != nil {
		return err
	}

	// 原子性地替换原文件
	return os.Rename(tempFilename, j.filename)
}

// GetAll 获取所有歌曲请求
func (j *JSONStore) GetAll() []SongRequest {
	j.mu.RLock()
	defer j.mu.RUnlock()

	result := make([]SongRequest, len(j.data))
	copy(result, j.data)
	return result
}

// GetByStatus 根据状态获取歌曲请求
func (j *JSONStore) GetByStatus(status string) []SongRequest {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var result []SongRequest
	for _, item := range j.data {
		if item.Status == status {
			result = append(result, item)
		}
	}
	return result
}

// Create 创建新的歌曲请求
func (j *JSONStore) Create(song SongRequest) SongRequest {
	j.mu.Lock()
	defer j.mu.Unlock()

	song.ID = j.nextID
	j.nextID++
	song.CreatedAt = time.Now()
	song.Status = "pending" // 默认状态为待播放

	j.data = append(j.data, song)

	// 异步保存到文件
	go j.saveToFile()

	return song
}

// UpdateStatus 更新歌曲状态
func (j *JSONStore) UpdateStatus(id uint, status string) (*SongRequest, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for i := range j.data {
		if j.data[i].ID == id {
			j.data[i].Status = status
			// 异步保存到文件
			go j.saveToFile()
			return &j.data[i], true
		}
	}
	return nil, false
}

// Delete 删除歌曲请求
func (j *JSONStore) Delete(id uint) bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	for i, item := range j.data {
		if item.ID == id {
			j.data = append(j.data[:i], j.data[i+1:]...)
			// 异步保存到文件
			go j.saveToFile()
			return true
		}
	}
	return false
}
