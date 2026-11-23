package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const grubsCommandOracle = "8173e160-9a22-43f7-90d1-9ada043bc1b7"

func TestGrubsCommandIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(grubsCommandOracle)
	if !ok || spec.Name != "Grub's Command" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// grubsCommandCast casts Grub's Command strictly, paying {3}{B}{R}.
func grubsCommandCast(t *testing.T, g *game.Game, modes []int, targets ...game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Grub's Command", TypeLine: "Kindred Sorcery — Goblin",
		OracleID: grubsCommandOracle, ManaCost: "{3}{B}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B06AddMana(me, "B", "R", "C", "C", "C")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: modes, Targets: targets})
}

// Copy + mill: a token copy of my Goblin; the target player mills five
// and every Goblin card among them goes to their hand, a deeper Goblin
// staying in the library.
func TestGrubsCommandCopiesAGoblinAndMillsGoblinsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gob := effectstest.B12Creature(g, me.ID, "Boggart Brute", "Creature — Goblin Warrior", 3, 2)
	deep := effectstest.SeedLibraryTop(opp, "Deep Goblin", "Creature — Goblin", "{R}")
	gobA := effectstest.SeedLibraryTop(opp, "Milled Goblin A", "Creature — Goblin", "{R}")
	elf := effectstest.SeedLibraryTop(opp, "Milled Elf", "Creature — Elf", "{G}")
	kindred := effectstest.SeedLibraryTop(opp, "Milled Kindred Goblin", "Kindred Instant — Goblin", "{R}")
	land := effectstest.SeedLibraryTop(opp, "Milled Mountain", "Basic Land — Mountain", "")
	bolt := effectstest.SeedLibraryTop(opp, "Milled Bolt", "Instant", "{R}")

	if err := grubsCommandCast(t, g, []int{0, 3},
		effectstest.ModeRef(game.TargetCard, gob, 0, 0),
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RfdCountNamed(g, me.ID, "Boggart Brute"); n != 2 {
		t.Errorf("%d Boggart Brutes, want the original and a token copy", n)
	}
	for _, id := range []uuid.UUID{gobA, kindred} {
		if !opp.Hand.Contains(id) {
			t.Error("a milled Goblin card is not in its owner's hand")
		}
	}
	for _, id := range []uuid.UUID{elf, land, bolt} {
		if !opp.Graveyard.Contains(id) {
			t.Error("a milled non-Goblin card left the graveyard")
		}
	}
	if !opp.Library.Contains(deep) {
		t.Error("the sixth card was milled")
	}
}

// Pump + destroy: the target player's creatures get +1/+1 and haste, and
// an artifact is destroyed.
func TestGrubsCommandPumpsAPlayersCreaturesAndDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := grubsCommandCast(t, g, []int{1, 2},
		effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0),
		effectstest.ModeRef(game.TargetCard, rock, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	c, _ := effectstest.BattlefieldCard(g, mine)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !game.HasKeyword(&c, "haste") {
		t.Errorf("my bear is %d/%d haste %v, want a 3/3 with haste", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "haste"))
	}
	o, _ := effectstest.BattlefieldCard(g, theirs)
	if o.CurrentPower() != 2 || game.HasKeyword(&o, "haste") {
		t.Error("the other player's creature was pumped")
	}
}

// The destroy bullet takes an artifact or creature only, the copy a
// Goblin you control, and one bullet alone is not "choose two".
func TestGrubsCommandRefusesBadTargetsOrOneMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := grubsCommandCast(t, g, []int{1, 2},
		effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0),
		effectstest.ModeRef(game.TargetCard, aura, 1, 0)); err == nil {
		t.Error("destroyed an enchantment")
	}
	if err := grubsCommandCast(t, g, []int{0, 1},
		effectstest.ModeRef(game.TargetCard, bear, 0, 0),
		effectstest.ModeRef(game.TargetPlayer, me.ID, 1, 0)); err == nil {
		t.Error("copied a non-Goblin")
	}
	if err := grubsCommandCast(t, g, []int{3}, effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0)); err == nil {
		t.Error("cast with one mode")
	}
}
