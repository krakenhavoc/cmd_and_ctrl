package ws

import (
	"log/slog"
	"os"
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// RoomManager is the keyed registry of active Rooms, one per active
// game. Replaces the S03 single-room singleton and is the boundary the
// hub consults on WebSocket upgrade to find the room a client is
// asking for (via /ws?game=<uuid>).
//
// At S04 rooms live for as long as the process does; Delete is exposed
// for future use but not wired into the lobby yet — if the server
// restarts, in-memory games are lost and clients get a "game not
// found" on reconnect. Persistence across restarts is deferred.
//
// Safe for concurrent use.
type RoomManager struct {
	mu      sync.RWMutex
	rooms   map[uuid.UUID]*Room
	log     *slog.Logger
	dumpDir string
}

// NewRoomManager constructs an empty RoomManager. dumpDir is the root
// passed to every Room created via Create; empty string disables
// crash-recovery writes for all rooms uniformly.
func NewRoomManager(log *slog.Logger, dumpDir string) *RoomManager {
	if log == nil {
		log = slog.Default()
	}
	return &RoomManager{
		rooms:   make(map[uuid.UUID]*Room),
		log:     log,
		dumpDir: dumpDir,
	}
}

// Create wraps the given game in a Room, registers it keyed by the
// game's ID, and returns the Room. If a room with the same game ID is
// already registered, the existing room is returned and no new Room is
// constructed — callers should treat the returned *Room as the
// canonical one.
func (m *RoomManager) Create(g *game.Game) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.rooms[g.ID]; ok {
		return existing
	}
	r := NewRoom(g, m.log, m.dumpDir)
	m.rooms[g.ID] = r
	return r
}

// Register inserts a pre-constructed Room into the manager. Used by
// tests and by any caller that needs a non-default Room configuration.
// Overwrites any existing room with the same game ID.
func (m *RoomManager) Register(r *Room) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rooms[r.Game.ID] = r
}

// Get returns the room for the given game ID, or nil if none is
// registered. A nil return is the manager's way of saying "unknown
// game" — callers should convert this into a wire-level error.
func (m *RoomManager) Get(id uuid.UUID) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[id]
}

// List returns a snapshot of all registered rooms, sorted by creation
// time of their underlying game (oldest first). The returned slice is
// safe to iterate without holding the manager's lock.
func (m *RoomManager) List() []*Room {
	m.mu.RLock()
	out := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		out = append(out, r)
	}
	m.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		return out[i].Game.CreatedAt.Before(out[j].Game.CreatedAt)
	})
	return out
}

// Delete removes a room from the manager and best-effort removes its
// on-disk artifacts (the crash-recovery snapshot and the replay log
// — without this, deleted games' files accumulate forever). Removal
// failures are logged, never fatal: the in-memory delete is the
// operation that matters. Does not close the room's underlying game
// or notify connected clients — the hub is responsible for evicting
// clients bound to the deleted game separately.
func (m *RoomManager) Delete(id uuid.UUID) {
	m.mu.Lock()
	delete(m.rooms, id)
	dumpDir := m.dumpDir
	m.mu.Unlock()

	if dumpDir == "" {
		return
	}
	// snapshotPath / replayPath (room.go) are the same helpers the
	// writers use, so the reaper can't drift from the layout.
	for _, path := range []string{
		snapshotPath(dumpDir, id),
		replayPath(dumpDir, id),
		// A deleted game must not be rebuilt by the next boot.
		restorePointPath(dumpDir, id),
	} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			m.log.Warn("delete game artifact failed", "path", path, "err", err)
		}
	}
}

// DumpDir returns the data root this manager writes artifacts under,
// or "" when disk persistence is disabled. Exposed so the lobby can
// store its own metadata beside the engine snapshots without a second
// configuration knob that could disagree with this one.
func (m *RoomManager) DumpDir() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dumpDir
}

// Count returns the number of registered rooms.
func (m *RoomManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// Singleton returns the sole registered room if exactly one exists,
// otherwise nil. Used by ServeWS as a backward-compat convenience so
// pre-S04 tests that dial /ws with no game query param still hit the
// one-and-only room. Multi-game deployments always pass ?game=...
// explicitly, so this is never reached in production.
func (m *RoomManager) Singleton() *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.rooms) != 1 {
		return nil
	}
	for _, r := range m.rooms {
		return r
	}
	return nil
}
