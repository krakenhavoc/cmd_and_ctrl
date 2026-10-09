package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// has_targets_test.go — #2853. Smart autopass stops on an opponent's
// stack item only for real interaction, and an activated ability is
// interaction only when it targets. `Move.HasTargets` is that signal:
// set on every announcement that chose a target, and never on a value
// ability like Mind Stone's draw or Evolving Wilds' search.

const (
	oracleMindStoneHT     = "c97361b5-af16-4a7b-af85-a429dbaf4ad2"
	oracleEvolvingWildsHT = "a75445d3-1303-4bb5-89ad-26ea93fecd48"
)

// The issue's board: three lands, Mind Stone and Evolving Wilds, plus a
// Grim Lavamancer with two cards in the graveyard and a Lightning Bolt
// in hand, so a targeted ability and a targeted cast sit beside the two
// untargeted ones in the same enumeration.
func TestHasTargetsIsSetOnlyOnMovesThatChooseATarget(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	for range 3 {
		battlefieldCard(g, seat, basic("Mountain", "Mountain"))
	}
	stone := battlefieldCard(g, seat, game.Card{
		Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}", OracleID: oracleMindStoneHT,
	})
	wilds := battlefieldCard(g, seat, game.Card{
		Name: "Evolving Wilds", TypeLine: "Land", OracleID: oracleEvolvingWildsHT,
	})
	lava := battlefieldCard(g, seat, game.Card{
		Name: "Grim Lavamancer", TypeLine: "Creature — Human Wizard", OracleID: oracleGrimLavamancer,
		ManaCost: "{R}", Power: 1, Toughness: 1,
	})
	for range 2 {
		graveyardCard(seat, game.Card{Name: "Fuel", TypeLine: "Sorcery", ManaCost: "{1}"})
	}
	bolt := handCard(seat, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})

	moves := legal.EnumerateFor(g, seat.ID)
	dispatchAll(t, g, seat.ID, moves)

	check := func(name string, source uuid.UUID, kind legal.Kind, want bool) {
		t.Helper()
		got := movesFrom(moves, source, kind)
		if len(got) == 0 {
			t.Fatalf("%s offered no %s move: %v", name, kind, labels(moves))
		}
		for _, m := range got {
			if m.HasTargets != want {
				t.Errorf("%s: has_targets = %v, want %v (%s)", name, m.HasTargets, want, m.Label)
			}
			if m.TargetsStack {
				t.Errorf("%s: targets_stack set on an empty stack (%s)", name, m.Label)
			}
		}
	}
	check("Mind Stone's draw", stone, legal.KindActivate, false)
	check("Evolving Wilds' search", wilds, legal.KindActivate, false)
	check("Grim Lavamancer's ping", lava, legal.KindActivate, true)
	check("Lightning Bolt", bolt, legal.KindCast, true)

	// Mana abilities never target and never carry the bit.
	for _, m := range moves {
		if m.Kind == legal.KindMana && m.HasTargets {
			t.Errorf("mana move %q has has_targets set", m.Label)
		}
	}
}

// A counterspell targets the stack, so it sets both bits.
func TestCounterspellSetsHasTargetsAndTargetsStack(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	clearHand(opp)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	cs := handCard(opp, game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", OracleID: oracleCounterspell})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	battlefieldCard(g, opp, basic("Island", "Island"))
	battlefieldCard(g, opp, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}

	moves := legal.EnumerateFor(g, opp.ID)
	got := movesFrom(moves, cs, legal.KindCast)
	if len(got) == 0 {
		t.Fatalf("Counterspell not offered: %v", labels(moves))
	}
	for _, m := range got {
		if !m.HasTargets || !m.TargetsStack {
			t.Errorf("Counterspell: has_targets=%v targets_stack=%v, want both (%s)", m.HasTargets, m.TargetsStack, m.Label)
		}
	}
}
