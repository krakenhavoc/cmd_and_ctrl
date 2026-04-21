package protocol

import (
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// view.go holds the wire-format projection of the server's
// authoritative game state. These types are what clients see; they
// are decoupled from the internal game package types so that the
// domain model can evolve without breaking the wire.
//
// The conversion functions (ViewOfGame, viewOfPlayer, etc.) are the
// single place where internal representation maps to the wire. When
// the game package grows new fields, they do NOT automatically show
// up on the wire until this file adds them — which is exactly the
// kind of explicit boundary we want.

// GameView is the complete JSON-serialisable snapshot of a Game.
type GameView struct {
	ID          string       `json:"id"`
	State       string       `json:"state"`
	Seats       []PlayerView `json:"seats"`
	Battlefield ZoneView     `json:"battlefield"`
	Stack       ZoneView     `json:"stack"`
	Exile       ZoneView     `json:"exile"`
	Turn        TurnView     `json:"turn"`
	// MulligansOpen reflects Game.MulligansOpen — true between Start
	// and the moment all seated players have committed to their
	// opening hand via the keep_hand action. Clients render the
	// keep / mulligan dialog while open. Added in S08.
	MulligansOpen bool `json:"mulligans_open"`
	// Monarch is the player ID currently designated as the monarch,
	// or empty string if no monarch is set. Sandbox marker; the must-
	// attack-when-able and combat-damage-transfer rules are not
	// enforced. Added in S10.
	Monarch string `json:"monarch,omitempty"`
	// Initiative is the player ID currently holding the initiative
	// (BG3 mechanic), or empty if unassigned. Same sandbox posture as
	// Monarch. Added in S10.
	Initiative string `json:"initiative,omitempty"`
	// Promises is the per-pair "I owe you" token tally as
	// "{from}->{to}" string keys → count. Zero entries are dropped on
	// the wire so the map stays small. Added in S10.
	Promises map[string]int `json:"promises,omitempty"`
	// Vote is the currently open vote (council's dilemma /
	// politics), or nil when no vote is in progress. Added in S10.
	Vote *VoteView `json:"vote,omitempty"`
	// UndoLimit is the per-player per-turn undo budget. Drives the
	// client's "undos remaining" indicator. Added in S11.
	UndoLimit int `json:"undo_limit,omitempty"`
	// StartingSeat is the seat index that took the first turn. Used
	// by the server to enforce the CR 103.7c turn-1 skip-draw rule;
	// surfaced on the wire so spectators / reconnects can render
	// "first player" UI affordances. Pre-S13 replays decode as 0,
	// which matches the only seat games started on before this field
	// existed. Added in S13.
	StartingSeat int `json:"starting_seat"`
	// StackItems is the announce-time metadata for every item
	// currently on the stack — caster, target list, modes, X,
	// distribution, hold-priority, split-second flags. Indexed in
	// stack order (bottom..top); the corresponding card (for spell
	// items) lives in the existing `stack` ZoneView. The dual
	// representation keeps existing zone plumbing intact while the
	// new cast/resolve UI reads StackItems for its rendering. Empty
	// when the stack is empty. Added in S13.1.
	StackItems []StackItemView `json:"stack_items,omitempty"`
	// PendingTriggers is the APNAP-ordered queue of triggered
	// abilities waiting to hit the stack (CR 603.3b). Drained on
	// next priority-grant boundary. Empty when no triggers are
	// pending. Added in S13.1.
	PendingTriggers []StackItemView `json:"pending_triggers,omitempty"`
	// SplitSecondActive mirrors `Game.SplitSecondActive` — true
	// while any item with split second is on the stack (CR 702.79).
	// Drives the client's "no responses allowed" UI gating. Added
	// in S13.1.
	SplitSecondActive bool `json:"split_second_active,omitempty"`
}

// StackItemView is the wire shape of a stack-item's announce-time
// metadata. Mirrors `game.StackItem` with UUIDs serialised as
// strings. See server/internal/game/stack.go for field semantics.
type StackItemView struct {
	ID           string          `json:"id"`
	Kind         string          `json:"kind"`
	Controller   string          `json:"controller"`
	Owner        string          `json:"owner"`
	SourceCardID string          `json:"source_card_id"`
	Label        string          `json:"label,omitempty"`
	Targets      []TargetRefView `json:"targets,omitempty"`
	Modes        []int           `json:"modes,omitempty"`
	XValue       int             `json:"x_value,omitempty"`
	Distribution map[string]int  `json:"distribution,omitempty"`
	HoldPriority bool            `json:"hold_priority,omitempty"`
	SplitSecond  bool            `json:"split_second,omitempty"`
}

// TargetRefView is the wire shape of a single announce-time target
// slot. Kind is one of "player", "card", "self", "none"; ID is the
// referenced UUID (zero string for "self" / "none"). See
// server/internal/game/stack.go for the canonical taxonomy.
type TargetRefView struct {
	Kind string `json:"kind"`
	ID   string `json:"id,omitempty"`
}

// VoteView is the wire form of game.Vote. Ballots is keyed by voter
// player ID (string UUID) for trivial JSON serialisation.
type VoteView struct {
	ID        string         `json:"id"`
	Topic     string         `json:"topic"`
	Options   []string       `json:"options"`
	Initiator string         `json:"initiator"`
	Ballots   map[string]int `json:"ballots"`
}

// PlayerView is the wire representation of a Player. Full-fidelity
// views come out of ViewOfGame; FilterViewFor then zeroes out any
// zones that should be hidden from a specific viewer (opponent hand
// cards, opponent library cards) while preserving the `count` so the
// UI can still render a placeholder stack.
type PlayerView struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Seat            int              `json:"seat"`
	Life            int              `json:"life"`
	Poison          int              `json:"poison,omitempty"`
	Energy          int              `json:"energy,omitempty"`
	Library         ZoneView         `json:"library"`
	Hand            ZoneView         `json:"hand"`
	Graveyard       ZoneView         `json:"graveyard"`
	Command         ZoneView         `json:"command"`
	CommanderDamage map[string]int   `json:"commander_damage"`
	LifeHistory     []LifeChangeView `json:"life_history"`
	// Eliminated reflects Player.Eliminated. Set when the player
	// concedes (S08); future state-based action work in S13+ may
	// also flip it. An eliminated player still appears in seats[],
	// still spectates, but the UI greys them out and disables their
	// quick-action buttons. Added in S08.
	Eliminated bool `json:"eliminated,omitempty"`
	// HandKept reflects Player.HandKept — true once the player has
	// committed to their opening hand. Drives the per-seat
	// "kept ✓" / "deciding…" indicator during the mulligan window.
	// Added in S08.
	HandKept bool `json:"hand_kept,omitempty"`
	// MulligansTaken reflects Player.MulligansTaken. Surfaced so the
	// UI can show "mulligans taken: N". Added in S08.
	MulligansTaken int `json:"mulligans_taken,omitempty"`
	// DeckImported reflects Player.DeckImported — true once the
	// player has had a real deck installed via ReplaceDeck (vs. the
	// 1-card placeholder commander handed out at AddPlayer time).
	// Drives the in-game deck-import modal in Game.svelte (S08.5
	// wave 1) — a card-count check is unreliable because the
	// placeholder also produces non-zero library / command counts.
	DeckImported bool `json:"deck_imported,omitempty"`
	// UndosRemaining is the per-turn undo budget left for this seat,
	// refreshed when the cursor enters their untap step. Drives the
	// "undos: N" indicator on the toolbar so the viewer can see at
	// a glance whether their next undo will be allowed. Added in S11.
	UndosRemaining int `json:"undos_remaining,omitempty"`

	// Discord identity (S12.5). Populated when the seat was claimed
	// via OAuth; zero values for manual-name joins. The client
	// builds /avatars/{discord_id}/{discord_avatar_hash}.png to
	// pull the cached portrait, and prefers display_name over name
	// for the seat label.
	DiscordID         string `json:"discord_id,omitempty"`
	DiscordAvatarHash string `json:"discord_avatar_hash,omitempty"`
	DisplayName       string `json:"display_name,omitempty"`

	// CommanderCasts is the per-commander cast count from the
	// command zone (S13.1, CR 903.8). Keyed by commander instance
	// UUID string. Drives the "+N tax" indicator next to the
	// commander tile. Sandbox: the engine doesn't enforce the
	// {2}-per-cast surcharge — players track mana themselves.
	// Omitted when empty.
	CommanderCasts map[string]int `json:"commander_casts,omitempty"`
}

// LifeChangeView is the wire representation of a single life-change
// log entry. Always emitted as part of PlayerView; the Player's
// canonical history is bounded server-side at MaxLifeHistoryEntries
// (S08), so the wire payload stays small without per-snapshot
// pruning here.
type LifeChangeView struct {
	Delta    int    `json:"delta"`
	NewTotal int    `json:"new_total"`
	At       string `json:"at"` // RFC3339
}

// ZoneView is the wire representation of a Zone. Count is sent
// explicitly so that future visibility-filtered views (e.g. "opponent
// library count but not contents") have a stable field to populate.
type ZoneView struct {
	Kind  string     `json:"kind"`
	Owner string     `json:"owner,omitempty"`
	Count int        `json:"count"`
	Cards []CardView `json:"cards"`
}

// CardView is the wire representation of a Card.
type CardView struct {
	InstanceID string `json:"instance_id"`
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	Controller string `json:"controller"`
	ScryfallID string `json:"scryfall_id,omitempty"`
	// TypeLine is Scryfall's printed type line ("Legendary Creature
	// — Human Wizard"). Carried so the client can filter "creatures
	// only" UIs (the combat panel) without a Scryfall round-trip.
	// Omitted for placeholder demo cards that have no resolved type.
	// Added in S08.
	TypeLine string `json:"type_line,omitempty"`
	// Power and Toughness are the parsed printed stats. Zero for
	// non-creatures and for any card with non-numeric printed stats
	// ("*", "1+*"). The client uses Power to label combat-panel
	// creature rows; ResolveCombatDamage uses CurrentPower (base +
	// counter modifiers) on the server side. Both omitempty for
	// non-creatures. Added in S08.
	Power       int            `json:"power,omitempty"`
	Toughness   int            `json:"toughness,omitempty"`
	Tapped      bool           `json:"tapped,omitempty"`
	Counters    map[string]int `json:"counters,omitempty"`
	IsCommander bool           `json:"is_commander,omitempty"`
	// DamageMarked is the damage currently noted on this creature
	// (S13.1 — feeds the lethal-damage SBA). Cleared by the
	// cleanup-step turn-based action and on zone exit. Only
	// meaningful for creatures on the battlefield; omitted when
	// zero. Added in S13.1.
	DamageMarked int `json:"damage_marked,omitempty"`
	// BattleX, BattleY are the normalised battlefield position in
	// [0, 1]. Emitted only for cards on the battlefield (other zones
	// clear them to zero on exit); clients should ignore these fields
	// outside the battlefield zone. omitempty drops them for cards that
	// have never been positioned (e.g. just-played cards waiting for a
	// drag-release).
	BattleX float64 `json:"battle_x,omitempty"`
	BattleY float64 `json:"battle_y,omitempty"`
	// AttackingTarget is the player ID this card is currently
	// declared to attack, or omitted if not declared. Cleared on
	// zone exit and by clear_combat. Added in S08.
	AttackingTarget string `json:"attacking_target,omitempty"`
	// BlockingTarget is the attacker instance ID this card is
	// currently declared to block, or omitted if not declared.
	// Cleared on zone exit and by clear_combat. Added in S08.
	BlockingTarget string `json:"blocking_target,omitempty"`
	// GoadedBy is the player ID who goaded this creature, or empty
	// when not goaded. Cleared on zone exit. Sandbox marker — the
	// must-attack-not-the-goader rule is not enforced. Added in S10.
	GoadedBy string `json:"goaded_by,omitempty"`
}

// TurnView is the wire representation of the turn cursor. PriorityHolder
// is the seat index that currently holds priority within the step
// (S07+); it equals ActiveSeat at every step boundary and rotates on
// pass_priority.
type TurnView struct {
	Number         int    `json:"number"`
	ActiveSeat     int    `json:"active_seat"`
	PriorityHolder int    `json:"priority_holder"`
	Phase          string `json:"phase"`
	Step           string `json:"step"`
}

// ViewOfGame builds a wire snapshot from a game.Game. It acquires a
// read lock on the game via ReadSnapshot and reads every field it
// needs in one consistent pass. Safe to call from any goroutine.
func ViewOfGame(g *game.Game) GameView {
	var view GameView
	g.ReadSnapshot(func() {
		view = GameView{
			ID:          g.ID.String(),
			State:       string(g.State),
			Seats:       viewOfSeats(g.Seats),
			Battlefield: viewOfZone(g.Battlefield),
			Stack:       viewOfZone(g.Stack),
			Exile:       viewOfZone(g.Exile),
			Turn: TurnView{
				Number:         g.Turn.Number,
				ActiveSeat:     g.Turn.ActiveSeat,
				PriorityHolder: g.Turn.PriorityHolder,
				Phase:          string(g.Turn.Phase),
				Step:           string(g.Turn.Step),
			},
			MulligansOpen:     g.MulligansOpen,
			Monarch:           uuidStringOrEmpty(g.Monarch),
			Initiative:        uuidStringOrEmpty(g.Initiative),
			Promises:          viewOfPromises(g.Promises),
			Vote:              viewOfVote(g.Vote),
			UndoLimit:         g.UndoLimit,
			StartingSeat:      g.StartingSeat,
			StackItems:        viewOfStackItemsInStackOrder(g),
			PendingTriggers:   viewOfStackItemSlice(g.PendingTriggers),
			SplitSecondActive: g.SplitSecondActive,
		}
	})
	return view
}

// viewOfStackItemsInStackOrder projects every stack item, ordered by
// the underlying Game.Stack zone (bottom..top). Spell items are
// matched to cards via InstanceID; ability items (which don't have a
// real card on the stack) are appended in StackMeta iteration order
// after the spell items. Returns nil for an empty stack so json
// omitempty drops the field.
func viewOfStackItemsInStackOrder(g *game.Game) []StackItemView {
	if len(g.StackMeta) == 0 {
		return nil
	}
	out := make([]StackItemView, 0, len(g.StackMeta))
	seen := make(map[uuid.UUID]bool, len(g.StackMeta))
	if g.Stack != nil {
		for _, c := range g.Stack.Cards {
			if item, ok := g.StackMeta[c.InstanceID]; ok && item != nil {
				out = append(out, viewOfStackItem(item))
				seen[item.ID] = true
			}
		}
	}
	for id, item := range g.StackMeta {
		if seen[id] || item == nil {
			continue
		}
		out = append(out, viewOfStackItem(item))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// viewOfStackItemSlice projects a slice of *StackItem (the pending-
// triggers queue) to the wire shape, preserving order. Returns nil
// for an empty input.
func viewOfStackItemSlice(items []*game.StackItem) []StackItemView {
	if len(items) == 0 {
		return nil
	}
	out := make([]StackItemView, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, viewOfStackItem(it))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// viewOfStackItem mirrors a single game.StackItem to its wire shape.
// Distribution and Targets are reallocated; scalar fields are
// stringified UUIDs.
func viewOfStackItem(it *game.StackItem) StackItemView {
	view := StackItemView{
		ID:           it.ID.String(),
		Kind:         string(it.Kind),
		Controller:   it.Controller.String(),
		Owner:        it.Owner.String(),
		SourceCardID: it.SourceCardID.String(),
		Label:        it.Label,
		Modes:        append([]int(nil), it.Modes...),
		XValue:       it.XValue,
		HoldPriority: it.HoldPriority,
		SplitSecond:  it.SplitSecond,
	}
	if len(it.Targets) > 0 {
		view.Targets = make([]TargetRefView, len(it.Targets))
		for i, t := range it.Targets {
			view.Targets[i] = TargetRefView{
				Kind: string(t.Kind),
				ID:   uuidStringOrEmpty(t.ID),
			}
		}
	}
	if len(it.Distribution) > 0 {
		view.Distribution = make(map[string]int, len(it.Distribution))
		for k, v := range it.Distribution {
			view.Distribution[k.String()] = v
		}
	}
	return view
}

func viewOfSeats(seats []*game.Player) []PlayerView {
	out := make([]PlayerView, len(seats))
	for i, p := range seats {
		out[i] = viewOfPlayer(p)
	}
	return out
}

func viewOfPlayer(p *game.Player) PlayerView {
	cmdrDamage := make(map[string]int, len(p.CommanderDamage))
	for k, v := range p.CommanderDamage {
		cmdrDamage[k.String()] = v
	}
	var cmdrCasts map[string]int
	if len(p.CommanderCasts) > 0 {
		cmdrCasts = make(map[string]int, len(p.CommanderCasts))
		for k, v := range p.CommanderCasts {
			cmdrCasts[k.String()] = v
		}
	}
	history := make([]LifeChangeView, len(p.LifeHistory))
	for i, c := range p.LifeHistory {
		history[i] = LifeChangeView{
			Delta:    c.Delta,
			NewTotal: c.NewTotal,
			At:       c.At.UTC().Format(time.RFC3339),
		}
	}
	return PlayerView{
		ID:                p.ID.String(),
		Name:              p.Name,
		Seat:              p.Seat,
		Life:              p.Life,
		Poison:            p.Poison,
		Energy:            p.Energy,
		Library:           viewOfZone(p.Library),
		Hand:              viewOfZone(p.Hand),
		Graveyard:         viewOfZone(p.Graveyard),
		Command:           viewOfZone(p.Command),
		CommanderDamage:   cmdrDamage,
		LifeHistory:       history,
		Eliminated:        p.Eliminated,
		HandKept:          p.HandKept,
		MulligansTaken:    p.MulligansTaken,
		DeckImported:      p.DeckImported,
		UndosRemaining:    p.UndosRemaining,
		DiscordID:         p.DiscordID,
		DiscordAvatarHash: p.DiscordAvatarHash,
		DisplayName:       p.DisplayName,
		CommanderCasts:    cmdrCasts,
	}
}

// viewOfPromises projects the per-pair promise map to wire format.
// Keys are encoded as "{from}->{to}" strings; entries with count <= 0
// are dropped so the wire payload stays sparse.
func viewOfPromises(in map[game.PromiseKey]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		if v <= 0 {
			continue
		}
		out[k.From.String()+"->"+k.To.String()] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// viewOfVote projects an open vote to wire format. Returns nil for a
// nil vote so json's omitempty drops the field entirely.
func viewOfVote(v *game.Vote) *VoteView {
	if v == nil {
		return nil
	}
	ballots := make(map[string]int, len(v.Ballots))
	for k, opt := range v.Ballots {
		ballots[k.String()] = opt
	}
	return &VoteView{
		ID:        v.ID.String(),
		Topic:     v.Topic,
		Options:   append([]string(nil), v.Options...),
		Initiator: v.Initiator.String(),
		Ballots:   ballots,
	}
}

// uuidStringOrEmpty returns u.String() unless u is the zero UUID, in
// which case it returns "" so json's omitempty drops the field. Used
// for nullable scalar IDs like Game.Monarch / Game.Initiative.
func uuidStringOrEmpty(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

func viewOfZone(z *game.Zone) ZoneView {
	if z == nil {
		return ZoneView{}
	}
	cards := make([]CardView, len(z.Cards))
	for i, c := range z.Cards {
		cards[i] = viewOfCard(c)
	}
	owner := ""
	if !z.IsShared() {
		owner = z.Owner.String()
	}
	return ZoneView{
		Kind:  string(z.Kind),
		Owner: owner,
		Count: len(z.Cards),
		Cards: cards,
	}
}

// FilterViewFor returns a copy of v with zones hidden from the given
// viewer zeroed out. The input is not mutated; only the copy's seat
// entries for non-viewer players get new (empty) card slices.
//
// Visibility rules at S04:
//   - Own seat: full fidelity (hand + library cards visible).
//   - Opponent seat: Hand.Cards and Library.Cards replaced with empty
//     slices; the Count field is preserved so the UI can render a
//     hidden stack. Graveyard and Command zones stay visible
//     (graveyard is public in MTG; command is public because
//     commanders are public).
//   - Shared zones (battlefield, stack, exile): unchanged.
//
// viewerID is the player UUID string; pass the empty string to get a
// "spectator" view where every opponent hand and library is hidden
// (i.e. no seat is treated as "own"). An observer without a claimed
// seat ends up here.
func FilterViewFor(v GameView, viewerID string) GameView {
	seats := make([]PlayerView, len(v.Seats))
	for i, p := range v.Seats {
		if p.ID != "" && p.ID == viewerID {
			seats[i] = p
			continue
		}
		hidden := p
		hidden.Hand = hideZoneContents(p.Hand)
		hidden.Library = hideZoneContents(p.Library)
		seats[i] = hidden
	}
	return GameView{
		ID:                v.ID,
		State:             v.State,
		Seats:             seats,
		Battlefield:       v.Battlefield,
		Stack:             v.Stack,
		Exile:             v.Exile,
		Turn:              v.Turn,
		MulligansOpen:     v.MulligansOpen,
		Monarch:           v.Monarch,
		Initiative:        v.Initiative,
		Promises:          v.Promises,
		Vote:              v.Vote,
		UndoLimit:         v.UndoLimit,
		StartingSeat:      v.StartingSeat,
		StackItems:        v.StackItems,
		PendingTriggers:   v.PendingTriggers,
		SplitSecondActive: v.SplitSecondActive,
	}
}

// hideZoneContents returns a copy of z with Cards replaced by an empty
// (but non-nil) slice. Non-nil matters: json.Marshal of a nil []CardView
// is `null`, but every other zone on the wire is `[]`; clients would
// see a shape inconsistency across seats.
func hideZoneContents(z ZoneView) ZoneView {
	return ZoneView{
		Kind:  z.Kind,
		Owner: z.Owner,
		Count: z.Count,
		Cards: []CardView{},
	}
}

func viewOfCard(c game.Card) CardView {
	var counters map[string]int
	if len(c.Counters) > 0 {
		counters = make(map[string]int, len(c.Counters))
		for k, v := range c.Counters {
			counters[k] = v
		}
	}
	view := CardView{
		InstanceID:   c.InstanceID.String(),
		Name:         c.Name,
		Owner:        c.Owner.String(),
		Controller:   c.Controller.String(),
		ScryfallID:   c.ScryfallID,
		TypeLine:     c.TypeLine,
		Power:        c.Power,
		Toughness:    c.Toughness,
		Tapped:       c.Tapped,
		Counters:     counters,
		IsCommander:  c.IsCommander,
		BattleX:      c.BattleX,
		BattleY:      c.BattleY,
		DamageMarked: c.DamageMarked,
	}
	if c.AttackingTarget != uuid.Nil {
		view.AttackingTarget = c.AttackingTarget.String()
	}
	if c.BlockingTarget != uuid.Nil {
		view.BlockingTarget = c.BlockingTarget.String()
	}
	if c.GoadedBy != uuid.Nil {
		view.GoadedBy = c.GoadedBy.String()
	}
	return view
}
