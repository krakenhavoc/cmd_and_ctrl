package game

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// settings.go is the table's configuration: the knobs a host may turn
// in the lobby or mid-game (ADR 0075 §2.2). It lives on the engine
// rather than on the lobby's GameMeta because the engine reads three
// of the fields at rules time (starting life at Start, the commander
// damage threshold in the SBA check, the undo budget at every untap),
// and because snapshots and undo would otherwise have to agree across
// two stores.
//
// WHO may change a setting is not decided here. Nothing in this file
// authorizes anything: the HTTP edge and the WebSocket action decide
// (ADR 0075 §2.1, lobby.CanManageTable). UpdateSettings records the
// actor it is handed so the log can name them, and that is all.

// UndoScope says whose undo entries a seat may take back.
type UndoScope string

const (
	// UndoScopeOwn is the long-standing rule: a seated player undoes
	// only their own entries. The server admin can undo anything.
	UndoScopeOwn UndoScope = "own"
	// UndoScopeHostAny also lets the host undo anyone's entry. Read by
	// the room layer once the host exists (ADR 0075 sub-PR 3).
	UndoScopeHostAny UndoScope = "host_any"
)

// BotPace is a named pacing preset for the AI seats (ADR 0033). The
// aiseat runner maps it to think times (ADR 0075 §2.2, sub-PR 6).
type BotPace string

const (
	BotPaceFast   BotPace = "fast"
	BotPaceNormal BotPace = "normal"
	BotPaceSlow   BotPace = "slow"
)

// UndoUnlimited is the UndoLimit value meaning "no per-turn budget".
// An unlimited budget is never debited and never refuses.
const UndoUnlimited = -1

// Setting bounds. Starting life and the commander damage threshold
// are house rules, so the range is wide; it only keeps out values no
// table means (a 0-life start, a threshold nothing can reach).
const (
	MinStartingLife    = 1
	MaxStartingLife    = 999
	MinCommanderDamage = 1
	MaxCommanderDamage = 99
)

// The settings keys. They are the snake_case wire names, and they are
// what EventSettingsChanged carries in Label so the log can say which
// knob moved.
const (
	SettingUndoLimit       = "undo_limit"
	SettingUndoScope       = "undo_scope"
	SettingStartingLife    = "starting_life"
	SettingCommanderDamage = "commander_damage"
	SettingBotPace         = "bot_pace"
	SettingAllowSpawn      = "allow_spawn"
)

// TableSettings is one table's configuration (ADR 0075 §2.2). The JSON
// tags are the snapshot's shape; the wire view has its own type.
type TableSettings struct {
	// UndoLimit is the per-player per-turn undo budget, refreshed on
	// each player's untap step. UndoUnlimited (-1) means no budget;
	// 0 means no undos, and is a real value — never rewritten to the
	// default.
	UndoLimit int `json:"undoLimit"`
	// UndoScope says whose entries a seat may undo.
	UndoScope UndoScope `json:"undoScope"`
	// StartingLife is each seat's life total at Start (40 in
	// Commander). Fixed once the game is active.
	StartingLife int `json:"startingLife"`
	// CommanderDamage is the combat damage from one commander that
	// loses the game (21 in Commander, CommanderDamageLethal). Read at
	// every state-based action check.
	CommanderDamage int `json:"commanderDamage"`
	// BotPace is the AI seats' pacing preset.
	BotPace BotPace `json:"botPace"`
	// AllowSpawn lets the host and admin spawn cards and tokens on a
	// live table, each spawn announced in the log (ADR 0075 §2.4).
	AllowSpawn bool `json:"allowSpawn"`
}

// DefaultTableSettings is what every new game starts with.
func DefaultTableSettings() TableSettings {
	return TableSettings{
		UndoLimit:       DefaultUndoLimit,
		UndoScope:       UndoScopeOwn,
		StartingLife:    StartingLife,
		CommanderDamage: CommanderDamageLethal,
		BotPace:         BotPaceNormal,
		AllowSpawn:      false,
	}
}

// SettingsPatch is a partial update: nil fields are left alone.
type SettingsPatch struct {
	UndoLimit       *int       `json:"undo_limit,omitempty"`
	UndoScope       *UndoScope `json:"undo_scope,omitempty"`
	StartingLife    *int       `json:"starting_life,omitempty"`
	CommanderDamage *int       `json:"commander_damage,omitempty"`
	BotPace         *BotPace   `json:"bot_pace,omitempty"`
	AllowSpawn      *bool      `json:"allow_spawn,omitempty"`
}

// Validate range-checks every present field. It does not know the
// game's state; UpdateSettings adds the StartingLife lock.
func (p SettingsPatch) Validate() error {
	if p.UndoLimit != nil && *p.UndoLimit < UndoUnlimited {
		return fmt.Errorf("%w: %s %d (want %d for unlimited, or 0 and up)",
			ErrInvalidSetting, SettingUndoLimit, *p.UndoLimit, UndoUnlimited)
	}
	if p.UndoScope != nil {
		switch *p.UndoScope {
		case UndoScopeOwn, UndoScopeHostAny:
		default:
			return fmt.Errorf("%w: %s %q", ErrInvalidSetting, SettingUndoScope, *p.UndoScope)
		}
	}
	if p.StartingLife != nil && (*p.StartingLife < MinStartingLife || *p.StartingLife > MaxStartingLife) {
		return fmt.Errorf("%w: %s %d (want %d..%d)",
			ErrInvalidSetting, SettingStartingLife, *p.StartingLife, MinStartingLife, MaxStartingLife)
	}
	if p.CommanderDamage != nil && (*p.CommanderDamage < MinCommanderDamage || *p.CommanderDamage > MaxCommanderDamage) {
		return fmt.Errorf("%w: %s %d (want %d..%d)",
			ErrInvalidSetting, SettingCommanderDamage, *p.CommanderDamage, MinCommanderDamage, MaxCommanderDamage)
	}
	if p.BotPace != nil {
		switch *p.BotPace {
		case BotPaceFast, BotPaceNormal, BotPaceSlow:
		default:
			return fmt.Errorf("%w: %s %q", ErrInvalidSetting, SettingBotPace, *p.BotPace)
		}
	}
	return nil
}

// UpdateSettings validates patch and applies its present fields, in
// the lobby or mid-game (ADR 0075 §2.3). actor is the player who
// asked, uuid.Nil for the server admin; it is recorded on the events,
// not checked.
//
// The whole patch is validated before anything is applied, so a patch
// with one bad field changes nothing. A StartingLife change after
// Start is refused with ErrStartingLifeLocked; a StartingLife in the
// lobby also resets every already-seated player's life, because that
// is the value Start would otherwise have stamped.
//
// An UndoLimit change refreshes every seat's UndosRemaining to the new
// value immediately. A CommanderDamage change is read at the next
// state-based action check, so lowering it below a player's damage
// from one commander loses them the game at that check.
//
// One EventSettingsChanged is emitted per field whose value actually
// changed, Label naming the setting and SettingOld / SettingNew
// carrying the values as text. A no-op patch emits nothing.
//
// Settings are not rolled back by undo: RestoreFrom carries the live
// value forward (see there).
func (g *Game) UpdateSettings(actor uuid.UUID, patch SettingsPatch) error {
	if err := patch.Validate(); err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.updateSettingsLocked(actor, patch)
}

func (g *Game) updateSettingsLocked(actor uuid.UUID, patch SettingsPatch) error {
	if patch.StartingLife != nil && *patch.StartingLife != g.Settings.StartingLife && g.State != StateLobby {
		return ErrStartingLifeLocked
	}
	old := g.Settings
	next := old
	if patch.UndoLimit != nil {
		next.UndoLimit = *patch.UndoLimit
	}
	if patch.UndoScope != nil {
		next.UndoScope = *patch.UndoScope
	}
	if patch.StartingLife != nil {
		next.StartingLife = *patch.StartingLife
	}
	if patch.CommanderDamage != nil {
		next.CommanderDamage = *patch.CommanderDamage
	}
	if patch.BotPace != nil {
		next.BotPace = *patch.BotPace
	}
	if patch.AllowSpawn != nil {
		next.AllowSpawn = *patch.AllowSpawn
	}
	g.Settings = next

	// The undo refresh runs whenever the patch names UndoLimit, even
	// at the same value — the legacy set_undo_limit action always
	// refreshed, and "set it to 1 again" is how a table hands every
	// seat its undo back.
	if patch.UndoLimit != nil {
		for _, p := range g.Seats {
			p.UndosRemaining = next.UndoLimit
		}
	}
	if g.State == StateLobby && next.StartingLife != old.StartingLife {
		for _, p := range g.Seats {
			p.Life = next.StartingLife
		}
	}

	emit := func(key, from, to string) {
		if from == to {
			return
		}
		g.EmitEvent(Event{
			Kind:       EventSettingsChanged,
			Actor:      actor,
			Label:      key,
			SettingOld: from,
			SettingNew: to,
		})
	}
	emit(SettingUndoLimit, strconv.Itoa(old.UndoLimit), strconv.Itoa(next.UndoLimit))
	emit(SettingUndoScope, string(old.UndoScope), string(next.UndoScope))
	emit(SettingStartingLife, strconv.Itoa(old.StartingLife), strconv.Itoa(next.StartingLife))
	emit(SettingCommanderDamage, strconv.Itoa(old.CommanderDamage), strconv.Itoa(next.CommanderDamage))
	emit(SettingBotPace, string(old.BotPace), string(next.BotPace))
	emit(SettingAllowSpawn, strconv.FormatBool(old.AllowSpawn), strconv.FormatBool(next.AllowSpawn))
	return nil
}

// TableSettingsSnapshot returns a copy of the current settings under
// the read lock.
func (g *Game) TableSettingsSnapshot() TableSettings {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Settings
}

// HasUndoBudget reports whether playerID may spend an undo now: true
// when the table's limit is UndoUnlimited, otherwise when the seat has
// UndosRemaining left. False for an unseated ID. The room layer asks
// this before restoring, so an undo is never half-applied.
func (g *Game) HasUndoBudget(playerID uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return false
	}
	return g.Settings.UndoLimit == UndoUnlimited || p.UndosRemaining > 0
}
