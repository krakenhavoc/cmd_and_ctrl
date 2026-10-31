package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// optional_behold_test.go — #4642: "As an additional cost to cast this
// spell, you may behold a <type>" is two kinds of cast — with the offer
// taken, naming the card it shows on reveal_ids, and without — and the
// beholding one is offered only to a seat that has a card to name. Every
// move is one the engine accepts (#544).

const (
	oracleHulksThunderclap  = "e377059d-a918-462f-920a-6b062bd5fcfd"
	oracleHyperionsAtomicVs = "e1e82283-01ec-447f-b0d4-5bd4619ab3d8"
)

type beholdMove struct {
	Optional []int
	Reveal   []string
	Targets  int
}

// beholdMovesOf decodes every cast of `src` the enumerator offered.
func beholdMovesOf(t *testing.T, moves []legal.Move, src uuid.UUID) []beholdMove {
	t.Helper()
	var out []beholdMove
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p struct {
			OptionalCosts []int             `json:"optional_costs"`
			RevealIDs     []string          `json:"reveal_ids"`
			Targets       []json.RawMessage `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		out = append(out, beholdMove{Optional: p.OptionalCosts, Reveal: p.RevealIDs, Targets: len(p.Targets)})
	}
	return out
}

func beholding(ms []beholdMove) []beholdMove {
	var out []beholdMove
	for _, m := range ms {
		if len(m.Optional) > 0 {
			out = append(out, m)
		}
	}
	return out
}

func unbeholding(ms []beholdMove) []beholdMove {
	var out []beholdMove
	for _, m := range ms {
		if len(m.Optional) == 0 {
			out = append(out, m)
		}
	}
	return out
}

// Hyperion's Atomic Vision with a Hero in hand: both forms are offered,
// the beholding one names the Hero card, and the plain one names nothing.
func TestOptionalBeholdOffersBothFormsWithAHeroInHand(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Plains")
	hero := handCard(seat, game.Card{Name: "Hero Friend", TypeLine: "Creature — Hero Human", ManaCost: "{W}"})
	vision := handCard(seat, game.Card{Name: "Hyperion's Atomic Vision", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: oracleHyperionsAtomicVs})
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	battlefieldCard(g, foe, game.Card{Name: "Tapped Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Tapped: true})

	moves := legal.EnumerateFor(g, seat.ID)
	got := beholdMovesOf(t, moves, vision)
	with, without := beholding(got), unbeholding(got)
	if len(with) == 0 || len(without) == 0 {
		t.Fatalf("moves = %+v, want a beholding and an unbeholding form", got)
	}
	for _, m := range with {
		if len(m.Optional) != 1 || m.Optional[0] != 0 || len(m.Reveal) != 1 || m.Reveal[0] != hero.String() {
			t.Errorf("beholding move = %+v, want offer 0 naming the Hero card %s", m, hero)
		}
	}
	for _, m := range without {
		if len(m.Reveal) != 0 {
			t.Errorf("unbeholding move = %+v, want no reveal_ids", m)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// A Hero already on the table is preferred to a card in hand: it shows
// nothing new.
func TestOptionalBeholdPrefersAPermanent(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Plains")
	handCard(seat, game.Card{Name: "Hero Friend", TypeLine: "Creature — Hero Human", ManaCost: "{W}"})
	onTable := battlefieldCard(g, seat, game.Card{Name: "Hero Scout", TypeLine: "Creature — Hero", Power: 1, Toughness: 1})
	vision := handCard(seat, game.Card{Name: "Hyperion's Atomic Vision", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: oracleHyperionsAtomicVs})
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	battlefieldCard(g, foe, game.Card{Name: "Tapped Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Tapped: true})

	moves := legal.EnumerateFor(g, seat.ID)
	with := beholding(beholdMovesOf(t, moves, vision))
	if len(with) == 0 || with[0].Reveal[0] != onTable.String() {
		t.Fatalf("beholding moves = %+v, want the permanent %s", with, onTable)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// With no Hero to behold only the plain form is offered: the seat is
// never offered a cast the engine would refuse.
func TestOptionalBeholdNotOfferedWithNothingToBehold(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Plains")
	handCard(seat, game.Card{Name: "Goblin Friend", TypeLine: "Creature — Goblin", ManaCost: "{R}"})
	vision := handCard(seat, game.Card{Name: "Hyperion's Atomic Vision", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: oracleHyperionsAtomicVs})
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	battlefieldCard(g, foe, game.Card{Name: "Tapped Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Tapped: true})

	moves := legal.EnumerateFor(g, seat.ID)
	got := beholdMovesOf(t, moves, vision)
	if len(got) == 0 || len(beholding(got)) != 0 {
		t.Fatalf("moves = %+v, want only the plain form", got)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Hulk's Thunderclap beholding a Gamma is announced with three targets,
// the plain form with two — and the beholding one needs a noncreature
// artifact or enchantment to name, so it is not offered without one.
func TestOptionalBeholdThunderclapThirdTarget(t *testing.T) {
	for _, withArtifact := range []bool{true, false} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		clearHand(seat)
		advanceTo(t, g, game.StepPrecombatMain)
		basicLands(g, seat, 2, "Forest")
		gamma := handCard(seat, game.Card{Name: "Gamma Friend", TypeLine: "Creature — Gamma Human", ManaCost: "{G}"})
		clap := handCard(seat, game.Card{Name: "Hulk's Thunderclap", TypeLine: "Sorcery", ManaCost: "{1}{G}", OracleID: oracleHulksThunderclap})
		battlefieldCard(g, seat, game.Card{Name: "Mine", TypeLine: "Creature — Test", Power: 2, Toughness: 2})
		battlefieldCard(g, foe, game.Card{Name: "Theirs", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
		if withArtifact {
			battlefieldCard(g, foe, game.Card{Name: "Pinger", TypeLine: "Artifact"})
		}

		moves := legal.EnumerateFor(g, seat.ID)
		got := beholdMovesOf(t, moves, clap)
		with, without := beholding(got), unbeholding(got)
		if len(without) == 0 {
			t.Fatalf("withArtifact=%v: no plain form in %+v", withArtifact, got)
		}
		for _, m := range without {
			if m.Targets != 2 {
				t.Errorf("withArtifact=%v: plain move names %d targets, want 2", withArtifact, m.Targets)
			}
		}
		if !withArtifact {
			if len(with) != 0 {
				t.Errorf("no artifact or enchantment: beholding forms %+v, want none", with)
			}
		} else {
			if len(with) == 0 {
				t.Fatalf("with an artifact: no beholding form in %+v", got)
			}
			for _, m := range with {
				if m.Targets != 3 || len(m.Reveal) != 1 || m.Reveal[0] != gamma.String() {
					t.Errorf("beholding move = %+v, want 3 targets and the Gamma card %s", m, gamma)
				}
			}
		}
		dispatchAll(t, g, seat.ID, moves)
	}
}
