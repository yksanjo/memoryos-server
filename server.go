package memoryos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// Server represents the MemoryOS HTTP server
type Server struct {
	memoryos *MemoryOS
	manager  *SharedMemoryManager
	addr     string
}

// NewServer creates a new MemoryOS server
func NewServer(memoryos *MemoryOS, addr string) *Server {
	return &Server{
		memoryos: memoryos,
		manager:  NewSharedMemoryManager(memoryos),
		addr:     addr,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/memory", s.handleMemory)
	http.HandleFunc("/memory/search", s.handleSearch)
	http.HandleFunc("/context", s.handleContext)
	http.HandleFunc("/agent", s.handleAgent)
	http.HandleFunc("/team", s.handleTeam)
	http.HandleFunc("/shared", s.handleShared)
	http.HandleFunc("/skill", s.handleSkill)
	http.HandleFunc("/stats", s.handleStats)

	log.Printf("MemoryOS server starting on %s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// ========== HEALTH ENDPOINT ==========

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health, err := s.manager.GetSystemHealth(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(health)
}

// ========== MEMORY ENDPOINTS ==========

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		s.storeMemory(w, r, ctx)
	case http.MethodGet:
		s.getMemory(w, r, ctx)
	case http.MethodDelete:
		s.deleteMemory(w, r, ctx)
	case http.MethodPut:
		s.updateMemory(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) storeMemory(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req struct {
		AgentID   string                 `json:"agent_id"`
		Type      string                 `json:"type"`
		Content   string                 `json:"content"`
		Metadata  map[string]interface{} `json:"metadata"`
		Tags      []string              `json:"tags"`
		Importance float64              `json:"importance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	memory := &Memory{
		AgentID:    req.AgentID,
		Type:       MemoryType(req.Type),
		Content:    req.Content,
		Metadata:   req.Metadata,
		Tags:       req.Tags,
		Importance: req.Importance,
	}

	if err := s.memoryos.StoreMemory(ctx, memory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"id": memory.ID})
}

func (s *Server) getMemory(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	agentID := r.URL.Query().Get("agent_id")
	memoryType := r.URL.Query().Get("type")
	memoryID := r.URL.Query().Get("id")

	if agentID == "" || memoryID == "" {
		http.Error(w, "agent_id and id required", http.StatusBadRequest)
		return
	}

	memory, err := s.memoryos.GetMemory(ctx, agentID, MemoryType(memoryType), memoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(memory)
}

func (s *Server) deleteMemory(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	agentID := r.URL.Query().Get("agent_id")
	memoryType := r.URL.Query().Get("type")
	memoryID := r.URL.Query().Get("id")

	if err := s.memoryos.DeleteMemory(ctx, agentID, MemoryType(memoryType), memoryID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (s *Server) updateMemory(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var memory Memory
	if err := json.NewDecoder(r.Body).Decode(&memory); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.memoryos.UpdateMemory(ctx, &memory); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// ========== SEARCH ENDPOINT ==========

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := r.URL.Query().Get("agent_id")
	query := r.URL.Query().Get("q")
	limit := 10

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	memories, err := s.memoryos.SearchMemories(ctx, agentID, query, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(memories)
}

// ========== CONTEXT ENDPOINT ==========

func (s *Server) handleContext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := r.URL.Query().Get("agent_id")
	maxTokens := 4000

	if tokensStr := r.URL.Query().Get("max_tokens"); tokensStr != "" {
		fmt.Sscanf(tokensStr, "%d", &maxTokens)
	}

	context, err := s.memoryos.GetContextWindow(ctx, agentID, maxTokens)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"context": context})
}

// ========== AGENT ENDPOINTS ==========

func (s *Server) handleAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		s.registerAgent(w, r, ctx)
	case http.MethodGet:
		s.getAgent(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) registerAgent(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var agent Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.manager.RegisterAgent(ctx, &agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(agent)
}

func (s *Server) getAgent(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	agentID := r.URL.Query().Get("id")

	agent, err := s.manager.GetAgent(ctx, agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(agent)
}

// ========== TEAM ENDPOINTS ==========

func (s *Server) handleTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		s.createTeam(w, r, ctx)
	case http.MethodGet:
		s.getTeam(w, r, ctx)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) createTeam(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var team Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.manager.CreateTeam(ctx, &team); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(team)
}

func (s *Server) getTeam(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	teamID := r.URL.Query().Get("id")

	team, err := s.manager.GetTeam(ctx, teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(team)
}

// ========== SHARED MEMORY ENDPOINTS ==========

func (s *Server) handleShared(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	path := r.URL.Path
	if strings.Contains(path, "/shared/value") {
		s.handleSharedValue(w, r, ctx)
		return
	}

	teamID := r.URL.Query().Get("team_id")
	key := r.URL.Query().Get("key")

	switch r.Method {
	case http.MethodPost:
		value := r.URL.Query().Get("value")
		if err := s.manager.CreateSharedValue(ctx, teamID, key, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	case http.MethodGet:
		value, err := s.manager.GetSharedValue(ctx, teamID, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"value": value})
	case http.MethodDelete:
		if err := s.manager.DeleteSharedValue(ctx, teamID, key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSharedValue(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	teamID := r.URL.Query().Get("team_id")
	key := r.URL.Query().Get("key")

	switch r.Method {
	case http.MethodPut:
		var req struct {
			Value string `json:"value"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if err := s.manager.UpdateSharedValue(ctx, teamID, key, req.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== SKILL ENDPOINTS ==========

func (s *Server) handleSkill(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	skillIndex := NewSkillIndex(s.memoryos)

	switch r.Method {
	case http.MethodPost:
		var skill Skill
		if err := json.NewDecoder(r.Body).Decode(&skill); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		skill.ID = uuid.New().String()
		if err := skillIndex.RegisterSkill(ctx, r.URL.Query().Get("agent_id"), &skill); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": skill.ID})
	case http.MethodGet:
		agentID := r.URL.Query().Get("agent_id")
		skillName := r.URL.Query().Get("name")

		if skillName != "" {
			skill, err := skillIndex.GetSkill(ctx, agentID, skillName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(skill)
		} else {
			skills, err := skillIndex.GetSkillsByCategory(ctx, agentID, r.URL.Query().Get("category"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(skills)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== STATS ENDPOINT ==========

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := r.URL.Query().Get("agent_id")

	stats, err := s.memoryos.GetMemoryStats(ctx, agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// ========== CLI STRUCTS ==========

// CLI represents the MemoryOS CLI
type CLI struct {
	memoryos *MemoryOS
	manager  *SharedMemoryManager
}

// NewCLI creates a new CLI
func NewCLI(memoryos *MemoryOS) *CLI {
	return &CLI{
		memoryos: memoryos,
		manager:  NewSharedMemoryManager(memoryos),
	}
}

// Run executes CLI commands
func (c *CLI) Run(args []string) error {
	if len(args) < 2 {
		return c.printHelp()
	}

	ctx := context.Background()
	command := args[1]

	switch command {
	case "store":
		return c.cmdStore(ctx, args[2:])
	case "get":
		return c.cmdGet(ctx, args[2:])
	case "search":
		return c.cmdSearch(ctx, args[2:])
	case "context":
		return c.cmdContext(ctx, args[2:])
	case "stats":
		return c.cmdStats(ctx, args[2:])
	case "agent":
		return c.cmdAgent(ctx, args[2:])
	case "team":
		return c.cmdTeam(ctx, args[2:])
	case "shared":
		return c.cmdShared(ctx, args[2:])
	case "skill":
		return c.cmdSkill(ctx, args[2:])
	case "help":
		return c.printHelp()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (c *CLI) cmdStore(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: store <agent_id> <type> <content>")
	}

	memory := &Memory{
		AgentID: args[0],
		Type:    MemoryType(args[1]),
		Content: strings.Join(args[2:], " "),
	}

	return c.memoryos.StoreMemory(ctx, memory)
}

func (c *CLI) cmdGet(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: get <agent_id> <type> <id>")
	}

	memory, err := c.memoryos.GetMemory(ctx, args[0], MemoryType(args[1]), args[2])
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(memory, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *CLI) cmdSearch(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: search <agent_id> <query>")
	}

	memories, err := c.memoryos.SearchMemories(ctx, args[0], args[1], 10)
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(memories, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *CLI) cmdContext(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: context <agent_id>")
	}

	context, err := c.memoryos.GetContextWindow(ctx, args[0], 4000)
	if err != nil {
		return err
	}

	fmt.Println(context)
	return nil
}

func (c *CLI) cmdStats(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stats <agent_id>")
	}

	stats, err := c.memoryos.GetMemoryStats(ctx, args[0])
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (c *CLI) cmdAgent(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: agent <name> [role]")
	}

	role := "worker"
	if len(args) > 1 {
		role = args[1]
	}

	agent := &Agent{
		Name: args[0],
		Role: role,
		Permissions: []string{"read", "write"},
		Metadata: make(map[string]interface{}),
	}

	return c.manager.RegisterAgent(ctx, agent)
}

func (c *CLI) cmdTeam(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: team <name>")
	}

	team := &Team{
		Name:        args[0],
		Description: "Created via CLI",
		Members:     []string{},
	}

	return c.manager.CreateTeam(ctx, team)
}

func (c *CLI) cmdShared(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: shared <team_id> <key> <value>")
	}

	return c.manager.CreateSharedValue(ctx, args[0], args[1], args[2])
}

func (c *CLI) cmdSkill(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: skill <agent_id> <name> <description>")
	}

	skill := &Skill{
		Name:        args[1],
		Description: strings.Join(args[2:], " "),
		Mastery:     0.0,
	}

	skillIndex := NewSkillIndex(c.memoryos)
	return skillIndex.RegisterSkill(ctx, args[0], skill)
}

func (c *CLI) printHelp() error {
	help := `
MemoryOS CLI - Redis for Agents

Usage:
  memoryos <command> [arguments]

Commands:
  store <agent_id> <type> <content>    Store a memory
  get <agent_id> <type> <id>           Get a memory
  search <agent_id> <query>            Search memories
  context <agent_id>                   Get context window
  stats <agent_id>                     Get memory statistics
  agent <name> [role]                  Register an agent
  team <name>                          Create a team
  shared <team_id> <key> <value>       Create shared value
  skill <agent_id> <name> <desc>       Register a skill
  help                                  Show this help

Types:
  episodic   - Event-based memories
  semantic   - Factual knowledge
  skill      - Procedural knowledge
  working    - Short-term context
  shared     - Multi-agent shared

Examples:
  memoryos store agent1 episodic "User asked about pricing"
  memoryos search agent1 "user interaction"
  memoryos context agent1
  memoryos stats agent1
`
	fmt.Println(help)
	return nil
}

// Main is the entry point for the CLI
func Main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: memoryos <command>")
		fmt.Println("Run 'memoryos help' for more information")
		return
	}

	// For demo purposes, create an in-memory version
	// In production, connect to actual Redis
	memoryos, err := NewMemoryOS(&MemoryOSConfig{
		RedisAddr: "localhost:6379",
		MaxTokens: 4000,
	})
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
		log.Println("Running in demo mode (no persistence)")
		return
	}

	cli := NewCLI(memoryos)
	if err := cli.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
