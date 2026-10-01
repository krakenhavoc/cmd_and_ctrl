package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// disturb_test.go — ADR 0107 §4 (#1855). A disturb card in its owner's
// graveyard is offered at its disturb cost and at no other, as its BACK
// face: the move's face, its target clause and its label are the back
// face's (CR 712.11a, 712.8c), and dispatchAll proves CastSpell accepts
// every one of them.

const (
	oracleBaithookAngler   = "c6bb4b41-8dae-429a-b928-ae9d39c74711"
	oracleDrogskolInfantry = "389bcb9f-4e66-4704-9968-a1c1574ec2c8"
)

// disturbCard is a disturb card as the importer builds it: a
// `transform` card front face up, with both faces' printed data.
func disturbCard(oracle string, front, back game.Face) game.Card {
	c := game.Card{OracleID: oracle, Layout: game.LayoutTransform, Faces: []game.Face{front, back}}
	c.SetFace(0)
	return c
}

func TestEnumeratorOffersDisturbAsTheBackFace(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Island")

	angler := graveyardCard(seat, disturbCard(oracleBaithookAngler,
		game.Face{Name: "Baithook Angler", TypeLine: "Creature — Human Peasant", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
		game.Face{Name: "Hook-Haunt Drifter", TypeLine: "Creature — Spirit", Colors: []string{"U"}, Power: 1, Toughness: 2},
	))

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, angler)
	if len(casts) != 1 {
		t.Fatalf("want exactly one cast of the disturb card, got %+v: %v", casts, labels(moves))
	}
	if casts[0].AlternativeCost != "disturb" || casts[0].FromZone != "graveyard" || casts[0].Face != 1 {
		t.Errorf("the cast is %+v, want disturb from the graveyard as face 1", casts[0])
	}
	if !hasLabel(moves, "Cast Hook-Haunt Drifter from graveyard (Disturb {1}{U})") {
		t.Errorf("the disturb move does not name the back face and the price: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// A disturbed Aura targets as it is cast (CR 303.4a) with the BACK
// face's clause; the front face, a creature, has no clause at all.
func TestEnumeratorTargetsADisturbedAuraWithTheBackFacesClause(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 4, "Plains")
	bear := battlefieldCard(g, seat, creature("Bear", "{1}{G}", 2, 2))

	infantry := graveyardCard(seat, disturbCard(oracleDrogskolInfantry,
		game.Face{Name: "Drogskol Infantry", TypeLine: "Creature — Spirit Soldier", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Drogskol Armaments", TypeLine: "Enchantment — Aura", Colors: []string{"W"}},
	))

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, infantry)
	if len(casts) == 0 {
		t.Fatalf("the disturbed Aura was not offered: %v", labels(moves))
	}
	sawBear := false
	for _, c := range casts {
		if c.AlternativeCost != "disturb" || c.Face != 1 {
			t.Errorf("cast %+v, want disturb as face 1", c)
		}
		if len(c.Targets) != 1 {
			t.Errorf("cast %+v names %d targets, want the one creature to enchant", c, len(c.Targets))
			continue
		}
		if c.Targets[0].ID == bear.String() {
			sawBear = true
		}
	}
	if !sawBear {
		t.Errorf("no disturb cast enchants the bear: %+v", casts)
	}
	dispatchAll(t, g, seat.ID, moves)
}
