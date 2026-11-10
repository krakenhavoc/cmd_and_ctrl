package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ashlingsCommandOracle = "9651d197-a408-4665-a292-58d5471f4ffc"

func TestAshlingsCommandIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(ashlingsCommandOracle)
	if !ok || spec.Name != "Ashling's Command" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// ashlingsCommandCast casts Ashling's Command strictly, paying {3}{U}{R}.
func ashlingsCommandCast(t *testing.T, g *game.Game, modes []int, targets ...game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ashling's Command", TypeLine: "Kindred Instant — Elemental",
		OracleID: ashlingsCommandOracle, ManaCost: "{3}{U}{R}", Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	b06AddMana(me, "U", "R", "C", "C", "C")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: modes, Targets: targets})
}

// Copy + damage: a token copy of my Elemental, and 2 damage to each
// creature the target player controls.
func TestAshlingsCommandCopiesAnElementalAndBurnsAPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flame := b12Creature(g, me.ID, "Flamekin Brawler", "Creature — Elemental Warrior", 3, 3)
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	if err := ashlingsCommandCast(t, g, []int{0, 2},
		modeRef(game.TargetCard, flame, 0, 0),
		modeRef(game.TargetPlayer, opp.ID, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := rfdCountNamed(g, me.ID, "Flamekin Brawler"); n != 2 {
		t.Errorf("%d Flamekin Brawlers, want the original and a token copy", n)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(mine) {
		t.Error("the 3/3 died, or a creature of another player was dealt damage")
	}
}

// Draw + Treasures: the target player draws two; the other target player
// creates two Treasures.
func TestAshlingsCommandDrawsForOnePlayerAndMakesTreasuresForAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "A", "B", "C")
	before := opp.Hand.Size()
	if err := ashlingsCommandCast(t, g, []int{1, 3},
		modeRef(game.TargetPlayer, opp.ID, 0, 0),
		modeRef(game.TargetPlayer, me.ID, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+2 {
		t.Errorf("their hand %d, want %d", got, before+2)
	}
	if n := rfdCountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
}

// The copy bullet needs an Elemental you control; one bullet alone is not
// "choose two".
func TestAshlingsCommandRefusesABadElementalTargetOrOneMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elemental", "Creature — Elemental", 2, 2)
	for name, target := range map[string]uuid.UUID{"a non-Elemental": bear, "an opponent's Elemental": theirs} {
		if err := ashlingsCommandCast(t, g, []int{0, 1},
			modeRef(game.TargetCard, target, 0, 0),
			modeRef(game.TargetPlayer, me.ID, 1, 0)); err == nil {
			t.Errorf("copied %s", name)
		}
	}
	if err := ashlingsCommandCast(t, g, []int{1}, modeRef(game.TargetPlayer, me.ID, 0, 0)); err == nil {
		t.Error("cast with one mode")
	}
}
