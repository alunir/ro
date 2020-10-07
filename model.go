package ro

// Model is an interface for redis objects
type Model interface {
	GetKeySuffix() string
	GetScoreMap() map[string]interface{}
	Serialized() []byte
	Deserialized([]byte)
	GetDatabaseNo() string
}
