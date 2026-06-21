// Package storage 為本機棋譜的持久化。
//
// Store 是語言中立的介面契約（存檔/載入/列表/刪除）；FileStore 為桌面/單機
// 的檔案系統實作，每局棋譜存為一個 xiangqi-record-v1 JSON 檔。行動端可另實作
// 平台儲存。跨語言一致性契約為棋譜本身的 JSON 格式（見 conformance/）。
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/yuanyu90221/xiangqi-platform/core/record"
)

// Entry 為棋局清單的輕量中介資料，免於載入完整走法即可顯示。
type Entry struct {
	ID     string
	Red    string
	Black  string
	Date   string
	Result string
}

// Store 定義本機棋譜持久化契約。
type Store interface {
	Save(id string, rec record.Record) error
	Load(id string) (record.Record, error)
	List() ([]Entry, error)
	Delete(id string) error
}

// FileStore 以單一目錄為後端：每個 id 對應 <dir>/<id>.json。
type FileStore struct {
	dir string
}

const ext = ".json"

// NewFileStore 建立以 dir 為後端的檔案儲存，目錄不存在則建立。
func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{dir: dir}, nil
}

// safePath 將 id 轉為儲存目錄內的檔案路徑；含路徑分隔或 .. 者拒絕，避免目錄穿越。
func (s *FileStore) safePath(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("storage: 空的識別字")
	}
	if id != filepath.Base(id) || strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return "", fmt.Errorf("storage: 不安全的識別字 %q", id)
	}
	return filepath.Join(s.dir, id+ext), nil
}

// Save 以 id 將棋譜存為 JSON 檔。
func (s *FileStore) Save(id string, rec record.Record) error {
	path, err := s.safePath(id)
	if err != nil {
		return err
	}
	data, err := record.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load 依 id 載回棋譜；不存在則回報錯誤。
func (s *FileStore) Load(id string) (record.Record, error) {
	path, err := s.safePath(id)
	if err != nil {
		return record.Record{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return record.Record{}, err
	}
	return record.Unmarshal(data)
}

// List 回傳已存棋譜的輕量 Entry，依 id 升冪排序。
func (s *FileStore) List() ([]Entry, error) {
	names, err := filepath.Glob(filepath.Join(s.dir, "*"+ext))
	if err != nil {
		return nil, err
	}
	slices.Sort(names)
	entries := make([]Entry, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		rec, err := record.Unmarshal(data)
		if err != nil {
			return nil, err
		}
		id := strings.TrimSuffix(filepath.Base(name), ext)
		entries = append(entries, Entry{
			ID: id, Red: rec.Red, Black: rec.Black, Date: rec.Date, Result: rec.Result,
		})
	}
	return entries, nil
}

// Delete 移除指定 id 的棋譜；不存在則回報錯誤。
func (s *FileStore) Delete(id string) error {
	path, err := s.safePath(id)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
