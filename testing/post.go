package rotesting

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/alunir/ro"
)

// Post is a test object
type Post struct {
	ID        uint64 `redis:"id"`
	Title     string `redis:"title"`
	Body      string `redis:"body"`
	UpdatedAt int64  `redis:"updated_at"`
}

// GetKeySuffix implements the types.Model interface
func (p *Post) GetKeySuffix() string {
	return fmt.Sprint(p.ID)
}

// GetScoreMap implements the types.Model interface
func (p *Post) GetScoreMap() map[string]interface{} {
	return map[string]interface{}{
		"id":     p.ID,
		"recent": p.UpdatedAt,
	}
}

// Serialized implements the types.Model interface
func (p *Post) Serialized() []byte {
	return []byte{}
}

// Deserialized implements the types.Model interface
func (p *Post) Deserialized(b []byte) {}

type Post_Serialized struct {
	ro.Model
	ID        uint64 `redis:"id"`
	UserID    uint64 `redis:"user_id"`
	Title     string `redis:"title"`
	Body      string `redis:"body"`
	CreatedAt int64  `redis:"created_at"`
}

func (p *Post_Serialized) GetKeySuffix() string {
	return fmt.Sprint(p.ID)
}

func (p *Post_Serialized) GetScoreMap() map[string]interface{} {
	return map[string]interface{}{
		"recent":                         p.CreatedAt,
		fmt.Sprintf("user:%d", p.UserID): p.CreatedAt,
	}
}

func (p *Post_Serialized) Serialized() []byte {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(p)
	if err != nil {
		panic("Failed to Serialized")
	}
	return buf.Bytes()
}

func (p *Post_Serialized) Deserialized(b []byte) {
	err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(p)
	if err != nil {
		panic("Failed to Deserialized. " + err.Error())
	}
}
