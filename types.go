package memoryos

import (
	"encoding/json"
	"time"
)

// MemoryType represents the type of memory stored
type MemoryType string

const (
	MemoryTypeEpisodic  MemoryType = "episodic"  // Event-based memories (experiences, interactions)
	MemoryTypeSemantic  MemoryType = "semantic"  // Factual knowledge (concepts, facts)
	MemoryTypeSkill     MemoryType = "skill"      // Procedural knowledge (how-to, capabilities)
	MemoryTypeWorking   MemoryType = "working"    // Short-term, immediate context
	MemoryTypeShared    MemoryType = "shared"    // Multi-agent shared memory
)

// Memory represents the base memory structure
type Memory struct {
	ID        string                 `json:"id"`
	Type      MemoryType             `json:"type"`
	AgentID   string                 `json:"agent_id"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Importance float64               `json:"importance"` // 0.0 - 1.0
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	AccessedAt time.Time             `json:"accessed_at"`
	AccessCount int                  `json:"access_count"`
	Tags       []string              `json:"tags,omitempty"`
	Embeddings []float64             `json:"embeddings,omitempty"`
}

// EpisodicMemory represents an event-based memory
type EpisodicMemory struct {
	Memory
	EventType   string                 `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Participants []string              `json:"participants,omitempty"`
	Emotion     string                 `json:"emotion,omitempty"`
	Outcome     string                 `json:"outcome,omitempty"`
	Lessons     []string               `json:"lessons,omitempty"`
}

// SemanticMemory represents factual/knowledge memories
type SemanticMemory struct {
	Memory
	Domain      string                 `json:"domain"`
	Concepts    []string               `json:"concepts"`
	Relations   map[string]string      `json:"relations,omitempty"`
	Confidence  float64                `json:"confidence"`
	Source      string                 `json:"source,omitempty"`
	Verified    bool                   `json:"verified"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// SkillMemory represents procedural/capability memories
type SkillMemory struct {
	Memory
	SkillName   string                 `json:"skill_name"`
	Category    string                 `json:"category"`
	Parameters  []string               `json:"parameters"`
	Returns     string                 `json:"returns"`
	Examples    []string               `json:"examples"`
	Prerequisites []string            `json:"prerequisites"`
	Mastery     float64                `json:"mastery"` // 0.0 - 1.0
	LastPracticed time.Time           `json:"last_practiced"`
	SuccessRate float64                `json:"success_rate"`
}

// WorkingMemory represents short-term context
type WorkingMemory struct {
	Memory
	Slot      int       `json:"slot"` // Position in context window
	TTL       time.Duration `json:"ttl"`
}

// SharedMemory represents multi-agent shared memory
type SharedMemory struct {
	Memory
	Scope      string                 `json:"scope"` // team, organization, global
	ACL        map[string]string      `json:"acl"`  // agent_id -> permission
	Version    int                   `json:"version"`
	Locked     bool                   `json:"locked"`
	LockOwner  string                 `json:"lock_owner,omitempty"`
}

// CompressedContext represents compressed context for LLM prompts
type CompressedContext struct {
	ID           string                 `json:"id"`
	AgentID      string                 `json:"agent_id"`
	OriginalSize int                    `json:"original_size"`
	CompressedSize int                  `json:"compressed_size"`
	Summary      string                 `json:"summary"`
	IncludedMemories []string           `json:"included_memories"`
	Strategy    string                 `json:"strategy"`
	CreatedAt    time.Time              `json:"created_at"`
}

// MemoryQuery represents a query for searching memories
type MemoryQuery struct {
	AgentID   string
	Type      *MemoryType
	Tags      []string
	Keywords  []string
	MinImportance float64
	Since     *time.Time
	Limit     int
	Offset    int
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	AgentID           string            `json:"agent_id"`
	TotalMemories     int               `json:"total_memories"`
	ByType            map[string]int    `json:"by_type"`
	TotalTokens       int               `json:"total_tokens"`
	AvgImportance     float64           `json:"avg_importance"`
	MostAccessed      []string          `json:"most_accessed"`
	LastConsolidation time.Time         `json:"last_consolidation"`
}

// ToJSON converts memory to JSON string
func (m *Memory) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	return string(data), err
}

// FromJSON creates memory from JSON string
func MemoryFromJSON(data string) (*Memory, error) {
	var m Memory
	err := json.Unmarshal([]byte(data), &m)
	return &m, err
}
