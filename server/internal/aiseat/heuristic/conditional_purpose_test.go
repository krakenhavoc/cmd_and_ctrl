package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// conditional_purpose_test.go — ADR 0126's amendment of 2026-10-15
// (#3762): what a card does only when its kicker is paid
// (`purpose.kicked`) or when it is turned face up
// (`purpose.turned_face_up`) is priced on the move that does that, and
// on no other.

// kickedWipe is Waterspout Elemental's kicked half: bounce every other
// creature and skip a turn.
func kickedWipe() *protocol.PurposeView {
	return &protocol.PurposeView{
		Sweep:      &protocol.SweepView{Matches: "creatures", How: "bounce", Partial: true},
		SkipsTurns: 1,
	}
}

func waterspout(id string, controller int, p *protocol.PurposeView) protocol.CardView {
	c := spell(id, controller, "Waterspout Elemental", "{3}{U}{U}")
	c.TypeLine = "Creature — Elemental"
	c.Power, c.Toughness = 3, 4
	c.Purpose = p
	return c
}

// castKicked is the cast move that pays optional cost 0, as the
// enumerator labels it.
func castKicked(t *testing.T, seat int, id, label string, optional ...int) legal.Move {
	m := castMove(t, seat, id, label)
	params := map[string]any{"instance_id": id, "from_zone": "hand"}
	if len(optional) > 0 {
		params["optional_costs"] = optional
	}
	m.Params = mustJSON(t, params)
	return m
}

func waterspoutBoard(hand protocol.CardView) protocol.GameView {
	bf := append(manaLands(7, 0, 100), bigArmy(0, 2, 200)...)
	bf = append(bf, bigArmy(1, 3, 300)...)
	return newView([]protocol.PlayerView{newSeat(0, withHand(hand)), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(6, 0, "precombat_main"))
}

// The unkicked cast is the creature it is: priced exactly as the same
// card declaring nothing, so it is not read as a mass bounce or a lost
// turn. The kicked cast is priced by what it adds.
func TestUnkickedCastIsNotPricedAsAWipeOrASkip(t *testing.T) {
	id := cardID(1)
	plain := waterspoutBoard(waterspout(id, 0, nil))
	cond := waterspoutBoard(waterspout(id, 0, &protocol.PurposeView{Kicked: kickedWipe()}))
	unkicked := castKicked(t, 0, id, "Cast Waterspout Elemental")
	kicked := castKicked(t, 0, id, "Cast Waterspout Elemental (Kicker {U})", 0)

	pol := heuristic.New()
	base := rankValue(t, pol, input(0, plain, passMove(0), unkicked), unkicked.Label)
	got := rankValue(t, pol, input(0, cond, passMove(0), unkicked, kicked), unkicked.Label)
	if !nearly(got, base) {
		t.Errorf("the unkicked cast priced %.3f, want the plain creature's %.3f", got, base)
	}

	// The kicked cast is priced as the same creature that declares the
	// wipe and the skip outright: nothing lost on the way through Kicked.
	flat := waterspoutBoard(waterspout(id, 0, kickedWipe()))
	want := rankValue(t, pol, input(0, flat, passMove(0), unkicked), unkicked.Label)
	if k := rankValue(t, pol, input(0, cond, passMove(0), unkicked, kicked), kicked.Label); !nearly(k, want) {
		t.Errorf("the kicked cast priced %.3f, want the unconditional declaration's %.3f", k, want)
	}
	if want >= base {
		t.Fatalf("fixture: the wipe and the skip (%.3f) should cost this board against the body (%.3f)", want, base)
	}
}

// Kicked and unkicked are two moves of one card, and the bot takes the
// one that is worth more: here the unkicked body, because bouncing its
// own two 4/4s and losing a turn is a bad trade.
func TestTheBotDoesNotKickAWipeThatHurtsIt(t *testing.T) {
	id := cardID(1)
	v := waterspoutBoard(waterspout(id, 0, &protocol.PurposeView{Kicked: kickedWipe()}))
	unkicked := castKicked(t, 0, id, "Cast Waterspout Elemental")
	kicked := castKicked(t, 0, id, "Cast Waterspout Elemental (Kicker {U})", 0)
	in := input(0, v, passMove(0), unkicked, kicked)
	pol := heuristic.New()
	if u, k := rankValue(t, pol, in, unkicked.Label), rankValue(t, pol, in, kicked.Label); k >= u {
		t.Errorf("kicked %.3f is no worse than unkicked %.3f", k, u)
	}
}

// A kicker the baseline does not price leaves the cast where it was:
// BaselineConfig prices neither sweeps nor turns.
func TestBaselineIgnoresTheKickedPurpose(t *testing.T) {
	id := cardID(1)
	plain := waterspoutBoard(waterspout(id, 0, nil))
	cond := waterspoutBoard(waterspout(id, 0, &protocol.PurposeView{Kicked: kickedWipe()}))
	kicked := castKicked(t, 0, id, "Cast Waterspout Elemental (Kicker {U})", 0)
	base := heuristic.NewWithConfig(heuristic.BaselineConfig())
	want := rankValue(t, base, input(0, plain, passMove(0), kicked), kicked.Label)
	if got := rankValue(t, base, input(0, cond, passMove(0), kicked), kicked.Label); !nearly(got, want) {
		t.Errorf("baseline priced the kicked cast %.3f, want %.3f", got, want)
	}
}

// facedown is a face-down permanent its controller may look at, carrying
// the one thing the view sends for it.
func facedown(id string, controller int, p *protocol.PurposeView) protocol.CardView {
	c := creature(id, controller, "Unstable Hulk", 2, 2)
	c.FaceDown, c.FaceDownKind, c.FaceVisible = true, "morphed", true
	c.Purpose = p
	return c
}

func turnUpMove(id string) legal.Move {
	raw := []byte(`{"card_id":"` + id + `","kind":"turn_face_up","route":"morph"}`)
	return legal.Move{
		Type: legal.TypeSpecialAction, Player: seatID(0), Kind: legal.KindSpecialAction,
		Label: "Turn face up {3}{R}{R} (morph)", Source: uuid.MustParse(id), Params: raw,
	}
}

// Turning Unstable Hulk up costs a turn, and the bot sees it: the flat
// special-action value less Weights.Turn. A permanent whose card
// declares nothing keeps the flat value.
func TestTurningAHulkUpIsChargedItsSkippedTurn(t *testing.T) {
	id := cardID(1)
	pol := heuristic.New()
	flat := pol.Config().SpecialActionValue
	turn := pol.Config().Weights.Turn
	if turn <= 0 {
		t.Fatalf("Weights.Turn = %v, want a positive price", turn)
	}
	view := func(p *protocol.PurposeView) protocol.GameView {
		return newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2), newSeat(3)},
			withBattlefield(append(manaLands(5, 0, 100), facedown(id, 0, p))...), withTurn(6, 0, "precombat_main"))
	}
	m := turnUpMove(id)
	declares := view(&protocol.PurposeView{TurnedFaceUp: &protocol.PurposeView{SkipsTurns: 1}})
	if got, want := rankValue(t, pol, input(0, declares, passMove(0), m), m.Label), flat-turn; !nearly(got, want) {
		t.Errorf("turning the Hulk up priced %.3f, want the flat %.3f less a turn: %.3f", got, flat, want)
	}
	silent := view(nil)
	if got := rankValue(t, pol, input(0, silent, passMove(0), m), m.Label); !nearly(got, flat) {
		t.Errorf("a face-down permanent that declares nothing priced %.3f, want the flat %.3f", got, flat)
	}
	base := heuristic.NewWithConfig(heuristic.BaselineConfig())
	if got := rankValue(t, base, input(0, declares, passMove(0), m), m.Label); !nearly(got, flat) {
		t.Errorf("baseline priced it %.3f, want the flat %.3f", got, flat)
	}
}
