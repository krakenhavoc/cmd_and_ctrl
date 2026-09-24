package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestScoreCountsEveryTermTheSprintNames(t *testing.T) {
	base := newView([]protocol.PlayerView{newSeat(0), newSeat(1)})
	me := seatID(0).String()
	ref := heuristic.Score(base, me)

	for _, tc := range []struct {
		name string
		view protocol.GameView
		up   bool
	}{
		{"life", newView([]protocol.PlayerView{newSeat(0, withLife(45)), newSeat(1)}), true},
		{"life lost", newView([]protocol.PlayerView{newSeat(0, withLife(12)), newSeat(1)}), false},
		{"hand", newView([]protocol.PlayerView{newSeat(0, withHandCount(4)), newSeat(1)}), true},
		{"creature", newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(creature(cardID(1), 0, "Bear", 2, 2))), true},
		{"non-creature permanent", newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(protocol.CardView{
				InstanceID: cardID(2), Controller: seatID(0).String(), Owner: seatID(0).String(),
				TypeLine: "Artifact", Name: "Rock", KnownByYou: true,
			})), true},
		{"untapped mana", newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(land(cardID(3), 0))), true},
		{"commander tax", newView([]protocol.PlayerView{newSeat(0, withCommanderCasts(3)), newSeat(1)}), false},
		{"opponent board", newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(creature(cardID(4), 1, "Bear", 2, 2))), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := heuristic.Score(tc.view, me)
			if tc.up && got <= ref {
				t.Errorf("score %.3f did not rise above the baseline %.3f", got, ref)
			}
			if !tc.up && got >= ref {
				t.Errorf("score %.3f did not fall below the baseline %.3f", got, ref)
			}
		})
	}
}

func TestUntappedManaBeatsTapped(t *testing.T) {
	me := seatID(0).String()
	up := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(land(cardID(1), 0)))
	down := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(land(cardID(1), 0, tapped())))
	if heuristic.Score(up, me) <= heuristic.Score(down, me) {
		t.Error("an untapped land must be worth more than a tapped one")
	}
}

func TestKeywordTableIsApplied(t *testing.T) {
	w := heuristic.DefaultWeights()
	vanilla := creature(cardID(1), 0, "Bear", 2, 2)
	for _, kw := range []string{"flying", "deathtouch", "lifelink", "trample", "double strike", "indestructible", "ward {2}", "prowess"} {
		flashy := creature(cardID(1), 0, "Bear", 2, 2, keywords(kw))
		if w.CreatureValue(&flashy) <= w.CreatureValue(&vanilla) {
			t.Errorf("%q did not raise the creature's value", kw)
		}
	}
	// #662: protection is counted off the PARSED projection, not off
	// the raw token — the policy package may not parse one. A token
	// with no projection behind it is a quality the engine's grammar
	// refused, and scoring it would price the creature above what it
	// does.
	pro := creature(cardID(1), 0, "Bear", 2, 2,
		keywords("protection from red"), protection(proColor("R", "red")))
	if w.CreatureValue(&pro) <= w.CreatureValue(&vanilla) {
		t.Error("protection from red did not raise the creature's value")
	}
	unparsed := creature(cardID(1), 0, "Bear", 2, 2, keywords("protection from monocolored"))
	if w.CreatureValue(&unparsed) != w.CreatureValue(&vanilla) {
		t.Error("a protection the engine does not enforce must not raise the creature's value")
	}
	// #706: prowess is cumulative (CR 702.108b), so a second instance
	// is worth a second bonus rather than a duplicate badge.
	one := creature(cardID(1), 0, "Monk", 1, 1, keywords("prowess"))
	two := creature(cardID(1), 0, "Monk", 1, 1, keywords("prowess", "prowess"))
	if w.CreatureValue(&two) <= w.CreatureValue(&one) {
		t.Error("a second instance of prowess did not raise the creature's value")
	}
	wall := creature(cardID(1), 0, "Wall", 0, 4, keywords("defender"))
	plain := creature(cardID(1), 0, "Plain", 0, 4)
	if w.CreatureValue(&wall) >= w.CreatureValue(&plain) {
		t.Error("defender must be a penalty, not a bonus")
	}
}

func TestLowLifeIsWorseThanLinear(t *testing.T) {
	me := seatID(0).String()
	at := func(n int) float64 {
		return heuristic.Score(newView([]protocol.PlayerView{newSeat(0, withLife(n)), newSeat(1)}), me)
	}
	// The drop from 8 to 4 must hurt more than the drop from 40 to 36.
	if at(36)-at(40) <= at(4)-at(8) {
		t.Errorf("life is priced linearly: 40→36 = %.3f, 8→4 = %.3f", at(36)-at(40), at(4)-at(8))
	}
}

func TestScoreIsRelativeToTheWholeTable(t *testing.T) {
	me := seatID(0).String()
	// Same board for me; the difference is entirely the opposition.
	alone := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(creature(cardID(1), 0, "Bear", 4, 4)))
	crowded := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(
			creature(cardID(1), 0, "Bear", 4, 4),
			creature(cardID(2), 1, "Bear", 4, 4),
			creature(cardID(3), 2, "Bear", 4, 4),
			creature(cardID(4), 3, "Bear", 4, 4),
		))
	if heuristic.Score(crowded, me) >= heuristic.Score(alone, me) {
		t.Error("three opponents matching my board must score worse than three empty seats")
	}
}

func TestThreatIsNotStrength(t *testing.T) {
	w := heuristic.DefaultWeights()
	// Two opponents with identical boards; one is at 8 life.
	healthy := &heuristic.SeatEval{ID: "a", Life: 38, Hand: 3, Board: 6}
	wounded := &heuristic.SeatEval{ID: "b", Life: 8, Hand: 3, Board: 6}
	if w.Threat(wounded) <= w.Threat(healthy) {
		t.Error("a wounded opponent must rank as the better place to point damage")
	}
	// And a big hand is scarier than an empty one.
	loaded := &heuristic.SeatEval{ID: "c", Life: 38, Hand: 7, Board: 6}
	if w.Threat(loaded) <= w.Threat(healthy) {
		t.Error("hidden cards must count toward the threat ranking")
	}
	// An eliminated seat is never a target.
	out := &heuristic.SeatEval{ID: "d", Life: 0, Hand: 0, Board: 40, Eliminated: true}
	if w.Threat(out) > 0 {
		t.Error("an eliminated seat must not rank as a threat")
	}
}

func TestEvaluateCountsBoardsPerSeat(t *testing.T) {
	w := heuristic.DefaultWeights()
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(1), 0, "Bear", 2, 2),
			creature(cardID(2), 0, "Sick", 3, 3, sick()),
			land(cardID(3), 0),
			land(cardID(4), 0, tapped()),
			creature(cardID(5), 1, "Theirs", 5, 5),
		))
	evals := w.Evaluate(v)
	mine := evals[seatID(0).String()]
	if mine.CreatureCount != 2 || mine.UntappedCreatures != 2 {
		t.Errorf("creature counts: %+v", mine)
	}
	if mine.UntappedMana != 1 {
		t.Errorf("untapped mana = %d, want 1 (the tapped land does not count)", mine.UntappedMana)
	}
	theirs := evals[seatID(1).String()]
	if theirs.CreatureCount != 1 || theirs.UntappedMana != 0 {
		t.Errorf("opponent breakdown: %+v", theirs)
	}
	if theirs.Creatures <= mine.Creatures {
		t.Errorf("one 5/5 should outvalue a 2/2 and a sick 3/3's creature total: %.2f vs %.2f",
			theirs.Creatures, mine.Creatures)
	}
}
