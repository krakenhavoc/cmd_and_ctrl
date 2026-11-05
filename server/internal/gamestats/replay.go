package gamestats

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ReplayFrames reads a replay JSONL (ADR 0041: one
// protocol.SnapshotPayload per committed action) as Frames, one frame
// decoded at a time. It is whole when every line decoded, every frame is
// gameID's (when gameID is set) and the last one is the ended game. A
// final line cut short by a crash makes it not whole, as does a stop.
//
// Each frame it yields carries only what the turn sampler reads: the
// game's id, state and turn, the newest log Seq, each seat's hand, land
// plays and elimination, and the battlefield's cards with their types,
// costs, tap and combat state and mana abilities. Libraries, graveyards,
// card text and the rest are skipped while decoding, which is most of a
// frame's bytes (a four-seat game's replay is 150 MB or more).
func ReplayFrames(r io.Reader, gameID string) Frames {
	return func(yield func(*protocol.GameView) bool) bool {
		dec := json.NewDecoder(r)
		var last string
		n := 0
		for {
			var snap leanSnapshot
			if err := dec.Decode(&snap); err != nil {
				return errors.Is(err, io.EOF) && n > 0 && last == "ended"
			}
			if gameID != "" && snap.Game.ID != gameID {
				return false
			}
			n++
			last = snap.Game.State
			v := snap.Game.view()
			if !yield(&v) {
				return false
			}
		}
	}
}

// ReplayFile is ReplayFrames over the file at path, opened when the
// frames are read. A file that cannot be opened yields nothing and is not
// whole.
func ReplayFile(path, gameID string) Frames {
	return func(yield func(*protocol.GameView) bool) bool {
		f, err := os.Open(path)
		if err != nil {
			return false
		}
		defer func() { _ = f.Close() }()
		return ReplayFrames(f, gameID)(yield)
	}
}

// The lean shapes below decode the part of a protocol.SnapshotPayload
// the sampler reads, under the same JSON names; encoding/json skips the
// rest without building it.

type leanSnapshot struct {
	Game leanView `json:"game"`
}

type leanView struct {
	ID          string       `json:"id"`
	State       string       `json:"state"`
	Seats       []leanPlayer `json:"seats"`
	Battlefield leanZone     `json:"battlefield"`
	Turn        leanTurn     `json:"turn"`
	Log         []leanLog    `json:"log"`
}

type leanTurn struct {
	Seq          int    `json:"seq"`
	Number       int    `json:"number"`
	ActiveSeat   int    `json:"active_seat"`
	Step         string `json:"step"`
	BetweenTurns bool   `json:"between_turns"`
	PhaseID      int    `json:"phase_id"`
	PhaseOrdinal int    `json:"phase_ordinal"`
}

type leanLog struct {
	Seq uint64 `json:"seq"`
}

type leanPlayer struct {
	ID                  string   `json:"id"`
	Seat                int      `json:"seat"`
	Eliminated          bool     `json:"eliminated"`
	Hand                leanZone `json:"hand"`
	LandsPlayedThisTurn int      `json:"lands_played_this_turn"`
	CantPlayLands       string   `json:"cant_play_lands"`
}

type leanZone struct {
	Cards []leanCard `json:"cards"`
}

type leanCard struct {
	InstanceID      string     `json:"instance_id"`
	Name            string     `json:"name"`
	Owner           string     `json:"owner"`
	Controller      string     `json:"controller"`
	ScryfallID      string     `json:"scryfall_id"`
	IsToken         bool       `json:"is_token"`
	TypeLine        string     `json:"type_line"`
	Tapped          bool       `json:"tapped"`
	PhasedOut       bool       `json:"phased_out"`
	SummoningSick   bool       `json:"summoning_sick"`
	AttackingTarget string     `json:"attacking_target"`
	ManaCost        string     `json:"mana_cost"`
	ManaAbilities   []leanMana `json:"mana_abilities"`
	Faces           []leanFace `json:"faces"`
}

type leanMana struct {
	Produced     string     `json:"produced"`
	ColorOptions [][]string `json:"color_options"`
	AddsNoMana   bool       `json:"adds_no_mana"`
}

type leanFace struct {
	TypeLine string `json:"type_line"`
}

func (v *leanView) view() protocol.GameView {
	out := protocol.GameView{
		ID:    v.ID,
		State: v.State,
		Turn: protocol.TurnView{Seq: v.Turn.Seq, Number: v.Turn.Number, ActiveSeat: v.Turn.ActiveSeat,
			Step: v.Turn.Step, BetweenTurns: v.Turn.BetweenTurns, PhaseID: v.Turn.PhaseID, PhaseOrdinal: v.Turn.PhaseOrdinal},
		Battlefield: protocol.ZoneView{Cards: cardViews(v.Battlefield.Cards)},
		Seats:       make([]protocol.PlayerView, len(v.Seats)),
	}
	var hi uint64
	for _, e := range v.Log {
		hi = max(hi, e.Seq)
	}
	if len(v.Log) > 0 {
		out.Log = []protocol.LogEvent{{Seq: hi}}
	}
	for i, p := range v.Seats {
		out.Seats[i] = protocol.PlayerView{ID: p.ID, Seat: p.Seat, Eliminated: p.Eliminated,
			Hand: protocol.ZoneView{Cards: cardViews(p.Hand.Cards)}, LandsPlayedThisTurn: p.LandsPlayedThisTurn,
			CantPlayLands: p.CantPlayLands}
	}
	return out
}

func cardViews(in []leanCard) []protocol.CardView {
	out := make([]protocol.CardView, len(in))
	for i, c := range in {
		out[i] = protocol.CardView{InstanceID: c.InstanceID, Name: c.Name, Owner: c.Owner, Controller: c.Controller,
			ScryfallID: c.ScryfallID, IsToken: c.IsToken, TypeLine: c.TypeLine, Tapped: c.Tapped, PhasedOut: c.PhasedOut,
			SummoningSick: c.SummoningSick, AttackingTarget: c.AttackingTarget, ManaCost: c.ManaCost}
		for _, m := range c.ManaAbilities {
			out[i].ManaAbilities = append(out[i].ManaAbilities,
				protocol.ManaAbilityView{Produced: m.Produced, ColorOptions: m.ColorOptions, AddsNoMana: m.AddsNoMana})
		}
		for _, f := range c.Faces {
			out[i].Faces = append(out[i].Faces, protocol.CardFaceView{TypeLine: f.TypeLine})
		}
	}
	return out
}
