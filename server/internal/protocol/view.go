package protocol

import (
	"time"

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
	InstanceID  string         `json:"instance_id"`
	Name        string         `json:"name"`
	Owner       string         `json:"owner"`
	Controller  string         `json:"controller"`
	ScryfallID  string         `json:"scryfall_id,omitempty"`
	Tapped      bool           `json:"tapped,omitempty"`
	Counters    map[string]int `json:"counters,omitempty"`
	IsCommander bool           `json:"is_commander,omitempty"`
	// BattleX, BattleY are the normalised battlefield position in
	// [0, 1]. Emitted only for cards on the battlefield (other zones
	// clear them to zero on exit); clients should ignore these fields
	// outside the battlefield zone. omitempty drops them for cards that
	// have never been positioned (e.g. just-played cards waiting for a
	// drag-release).
	BattleX float64 `json:"battle_x,omitempty"`
	BattleY float64 `json:"battle_y,omitempty"`
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
		}
	})
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
	history := make([]LifeChangeView, len(p.LifeHistory))
	for i, c := range p.LifeHistory {
		history[i] = LifeChangeView{
			Delta:    c.Delta,
			NewTotal: c.NewTotal,
			At:       c.At.UTC().Format(time.RFC3339),
		}
	}
	return PlayerView{
		ID:              p.ID.String(),
		Name:            p.Name,
		Seat:            p.Seat,
		Life:            p.Life,
		Poison:          p.Poison,
		Energy:          p.Energy,
		Library:         viewOfZone(p.Library),
		Hand:            viewOfZone(p.Hand),
		Graveyard:       viewOfZone(p.Graveyard),
		Command:         viewOfZone(p.Command),
		CommanderDamage: cmdrDamage,
		LifeHistory:     history,
		Eliminated:      p.Eliminated,
	}
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
		ID:          v.ID,
		State:       v.State,
		Seats:       seats,
		Battlefield: v.Battlefield,
		Stack:       v.Stack,
		Exile:       v.Exile,
		Turn:        v.Turn,
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
	return CardView{
		InstanceID:  c.InstanceID.String(),
		Name:        c.Name,
		Owner:       c.Owner.String(),
		Controller:  c.Controller.String(),
		ScryfallID:  c.ScryfallID,
		Tapped:      c.Tapped,
		Counters:    counters,
		IsCommander: c.IsCommander,
		BattleX:     c.BattleX,
		BattleY:     c.BattleY,
	}
}
