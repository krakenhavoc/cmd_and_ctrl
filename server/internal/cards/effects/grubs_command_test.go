package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const grubsCommandOracle = "8173e160-9a22-43f7-90d1-9ada043bc1b7"

func TestGrubsCommandIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(grubsCommandOracle)
	if !ok || spec.Name != "Grub's Command" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
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
	advanceTo(t, g, game.StepPrecombatMain)
	b06AddMana(me, "B", "R", "C", "C", "C")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: modes, Targets: targets})
}

// Copy + mill: a token copy of my Goblin; the target player mills five
// and every Goblin card among them goes to their hand, a deeper Goblin
// staying in the library.
func TestGrubsCommandCopiesAGoblinAndMillsGoblinsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gob := b12Creature(g, me.ID, "Boggart Brute", "Creature — Goblin Warrior", 3, 2)
	deep := seedLibraryTop(opp, "Deep Goblin", "Creature — Goblin", "{R}")
	gobA := seedLibraryTop(opp, "Milled Goblin A", "Creature — Goblin", "{R}")
	elf := seedLibraryTop(opp, "Milled Elf", "Creature — Elf", "{G}")
	kindred := seedLibraryTop(opp, "Milled Kindred Goblin", "Kindred Instant — Goblin", "{R}")
	land := seedLibraryTop(opp, "Milled Mountain", "Basic Land — Mountain", "")
	bolt := seedLibraryTop(opp, "Milled Bolt", "Instant", "{R}")

	if err := grubsCommandCast(t, g, []int{0, 3},
		modeRef(game.TargetCard, gob, 0, 0),
		modeRef(game.TargetPlayer, opp.ID, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := rfdCountNamed(g, me.ID, "Boggart Brute"); n != 2 {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := grubsCommandCast(t, g, []int{1, 2},
		modeRef(game.TargetPlayer, me.ID, 0, 0),
		modeRef(game.TargetCard, rock, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	c, _ := battlefieldCard(g, mine)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !game.HasKeyword(&c, "haste") {
		t.Errorf("my bear is %d/%d haste %v, want a 3/3 with haste", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "haste"))
	}
	o, _ := battlefieldCard(g, theirs)
	if o.CurrentPower() != 2 || game.HasKeyword(&o, "haste") {
		t.Error("the other player's creature was pumped")
	}
}

// The destroy bullet takes an artifact or creature only, the copy a
// Goblin you control, and one bullet alone is not "choose two".
func TestGrubsCommandRefusesBadTargetsOrOneMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := grubsCommandCast(t, g, []int{1, 2},
		modeRef(game.TargetPlayer, me.ID, 0, 0),
		modeRef(game.TargetCard, aura, 1, 0)); err == nil {
		t.Error("destroyed an enchantment")
	}
	if err := grubsCommandCast(t, g, []int{0, 1},
		modeRef(game.TargetCard, bear, 0, 0),
		modeRef(game.TargetPlayer, me.ID, 1, 0)); err == nil {
		t.Error("copied a non-Goblin")
	}
	if err := grubsCommandCast(t, g, []int{3}, modeRef(game.TargetPlayer, opp.ID, 0, 0)); err == nil {
		t.Error("cast with one mode")
	}
}
