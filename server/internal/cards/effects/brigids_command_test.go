package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const brigidsCommandOracle = "6487fc1b-5c83-4d09-94d3-175725bcb922"

func TestBrigidsCommandIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(brigidsCommandOracle)
	if !ok || spec.Name != "Brigid's Command" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// brigidsCommandCast casts Brigid's Command strictly, paying {1}{G}{W}.
func brigidsCommandCast(t *testing.T, g *game.Game, modes []int, targets ...game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Brigid's Command", TypeLine: "Kindred Sorcery — Kithkin",
		OracleID: brigidsCommandOracle, ManaCost: "{1}{G}{W}", Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	b06AddMana(me, "G", "W", "C")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: modes, Targets: targets})
}

// Copy + token: a token copy of my Kithkin, and a Kithkin token for the
// target player (an opponent here, so it is theirs).
func TestBrigidsCommandCopiesAKithkinAndGivesAPlayerAKithkin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kith := b12Creature(g, me.ID, "Kithkin Daggerdare", "Creature — Kithkin Soldier", 2, 2)
	if err := brigidsCommandCast(t, g, []int{0, 1},
		modeRef(game.TargetCard, kith, 0, 0),
		modeRef(game.TargetPlayer, opp.ID, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := rfdCountNamed(g, me.ID, "Kithkin Daggerdare"); n != 2 {
		t.Errorf("%d Daggerdares, want the original and a token copy", n)
	}
	if n := rfdCountNamed(g, opp.ID, "Kithkin"); n != 1 {
		t.Errorf("opponent has %d Kithkin tokens, want 1", n)
	}
	if n := rfdCountNamed(g, me.ID, "Kithkin"); n != 0 {
		t.Errorf("caster has %d Kithkin tokens, want 0", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Kithkin" && (c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 2) {
			t.Errorf("Kithkin token = %d/%d %v, want a 1/1 green and white", c.Power, c.Toughness, c.Colors)
		}
	}
}

// Pump + fight: the bullets run in order, so my 2/2 fights as a 5/5 and
// survives a 4/4 it kills.
func TestBrigidsCommandPumpsThenFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Beast", "Creature — Beast", 4, 4)
	if err := brigidsCommandCast(t, g, []int{2, 3},
		modeRef(game.TargetCard, mine, 0, 0),
		modeRef(game.TargetCard, mine, 1, 0),
		modeRef(game.TargetCard, theirs, 1, 1)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4/4 survived a fight with a 5/5")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my pumped creature died to 4 damage")
	}
}

// The copy bullet needs a Kithkin you control, the fight's second target
// an opponent's creature, and one bullet alone is not "choose two".
func TestBrigidsCommandRefusesBadTargetsOrOneMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirKith := b12Creature(g, opp.ID, "Their Kithkin", "Creature — Kithkin", 2, 2)
	for name, target := range map[string]uuid.UUID{"a non-Kithkin": bear, "an opponent's Kithkin": theirKith} {
		if err := brigidsCommandCast(t, g, []int{0, 1},
			modeRef(game.TargetCard, target, 0, 0),
			modeRef(game.TargetPlayer, me.ID, 1, 0)); err == nil {
			t.Errorf("copied %s", name)
		}
	}
	other := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	if err := brigidsCommandCast(t, g, []int{1, 3},
		modeRef(game.TargetPlayer, me.ID, 0, 0),
		modeRef(game.TargetCard, bear, 1, 0),
		modeRef(game.TargetCard, other, 1, 1)); err == nil {
		t.Error("fought my own creature")
	}
	if err := brigidsCommandCast(t, g, []int{1}, modeRef(game.TargetPlayer, me.ID, 0, 0)); err == nil {
		t.Error("cast with one mode")
	}
}
