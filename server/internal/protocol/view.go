package protocol

import (
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

// PlayerView is the wire representation of a Player. S03 sends every
// player's full hand contents — there is no visibility filtering yet.
// Visibility (hide opponent hands, show counts only) arrives in S04
// with authentication, since "whose view is this" is only meaningful
// once we know who's asking.
type PlayerView struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Seat            int            `json:"seat"`
	Life            int            `json:"life"`
	Poison          int            `json:"poison,omitempty"`
	Energy          int            `json:"energy,omitempty"`
	Library         ZoneView       `json:"library"`
	Hand            ZoneView       `json:"hand"`
	Graveyard       ZoneView       `json:"graveyard"`
	Command         ZoneView       `json:"command"`
	CommanderDamage map[string]int `json:"commander_damage"`
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
	Tapped      bool           `json:"tapped,omitempty"`
	Counters    map[string]int `json:"counters,omitempty"`
	IsCommander bool           `json:"is_commander,omitempty"`
}

// TurnView is the wire representation of the turn cursor.
type TurnView struct {
	Number     int    `json:"number"`
	ActiveSeat int    `json:"active_seat"`
	Phase      string `json:"phase"`
	Step       string `json:"step"`
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
				Number:     g.Turn.Number,
				ActiveSeat: g.Turn.ActiveSeat,
				Phase:      string(g.Turn.Phase),
				Step:       string(g.Turn.Step),
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
		Tapped:      c.Tapped,
		Counters:    counters,
		IsCommander: c.IsCommander,
	}
}
