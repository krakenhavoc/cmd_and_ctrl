package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCelebrationCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.ArmoryMiceOracle:           "Armory Mice",
		effectstest.AshPartyCrasherOracle:      "Ash, Party Crasher",
		effectstest.BelligerentOfTheBallOracle: "Belligerent of the Ball",
		effectstest.BespokeBattlegarbOracle:    "Bespoke Battlegarb",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// celebrationCountFor is the seam's count, read under the lock.
func celebrationCountFor(g *game.Game, you uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.NonlandPermanentsEnteredThisTurn(you) })
	return n
}

func celebrationToughnessFor(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.CardOnBattlefield(g, id)
	if !ok {
		t.Fatalf("card %s not on battlefield", id)
	}
	return c.CurrentToughness()
}

// --- the condition helper ---------------------------------------------------

// Celebration is two or more: one entry is not enough, a land does not
// help, and an opponent's entries are theirs.
func TestCelebrationNeedsTwoNonlandPermanentsOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	holds := func() bool {
		var ok bool
		g.WithWriteLock(func() { ok = effects.Celebration(g, me.ID) })
		return ok
	}
	if holds() {
		t.Fatal("celebration before anything entered")
	}
	effectstest.CelebrationTokens(t, g, me.ID, 1)
	effectstest.CelebrationEnterFromHand(t, g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.CelebrationTokens(t, g, opp.ID, 3)
	if holds() {
		t.Errorf("celebration with one token, a land and an opponent's tokens (count %d)", celebrationCountFor(g, me.ID))
	}
	effectstest.CelebrationEnterFromHand(t, g, me.ID, "Ornithopter", "Artifact Creature — Thopter", "")
	if !holds() {
		t.Errorf("no celebration after a token and an artifact creature (count %d)", celebrationCountFor(g, me.ID))
	}
}

// --- Armory Mice ---------------------------------------------------------------

func TestArmoryMiceGetsPlusZeroPlusTwoWhileCelebrating(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mice := effectstest.PushCatalogPermanent(g, me.ID, "Armory Mice", "Creature — Mouse", effectstest.ArmoryMiceOracle, false)
	base := celebrationCountFor(g, me.ID)
	if base >= 2 {
		t.Fatalf("setup: already %d entries", base)
	}
	if got := celebrationToughnessFor(t, g, mice); got != 1 {
		t.Fatalf("toughness %d with no celebration, want 1", got)
	}
	effectstest.CelebrationTokens(t, g, me.ID, 2-base)
	if got := celebrationToughnessFor(t, g, mice); got != 3 {
		t.Errorf("toughness %d while celebrating, want 3", got)
	}
	if got := effectstest.CurrentPowerFor(t, g, mice); got != 1 {
		t.Errorf("power %d while celebrating, want 1 (+0/+2)", got)
	}
	// The tokens leaving does not end it: they entered this turn.
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == effectstest.CelebrationSoldierTokenName {
				if err := g.SacrificePermanentForEffect(c.InstanceID); err != nil {
					t.Fatalf("sacrifice: %v", err)
				}
			}
		}
	})
	if got := celebrationToughnessFor(t, g, mice); got != 3 {
		t.Errorf("toughness %d after the tokens left, want 3", got)
	}
	// It ends with the turn.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if got := celebrationToughnessFor(t, g, mice); got != 1 {
		t.Errorf("toughness %d on the next turn, want 1", got)
	}
}

// The Mice count themselves: the Mice and one other nonland permanent
// entering is celebration; the Mice and a land is not.
func TestArmoryMiceCountsItselfEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mice := effectstest.CelebrationEnterFromHand(t, g, me.ID, "Armory Mice", "Creature — Mouse", effectstest.ArmoryMiceOracle)
	effectstest.CelebrationEnterFromHand(t, g, me.ID, "Forest", "Basic Land — Forest", "")
	if got := celebrationToughnessFor(t, g, mice); got != 1 {
		t.Errorf("the Mice and a land: toughness %d, want 1", got)
	}
	effectstest.CelebrationTokens(t, g, me.ID, 1)
	if got := celebrationToughnessFor(t, g, mice); got != 3 {
		t.Errorf("the Mice and a token: toughness %d, want 3", got)
	}
}

// --- Ash, Party Crasher ------------------------------------------------------------

func TestAshPartyCrasherGetsACounterWhenAttackingWhileCelebrating(t *testing.T) {
	for _, celebrate := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ash := effectstest.PushCatalogPermanent(g, me.ID, "Ash, Party Crasher", "Legendary Creature — Human Peasant", effectstest.AshPartyCrasherOracle, false)
		if celebrate {
			effectstest.CelebrationTokens(t, g, me.ID, 2)
		}
		effectstest.DeclareAttack(t, g, opp.ID, ash)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if celebrate {
			want = 1
		}
		if got := effectstest.CounterCount(g, ash, game.CounterPlusOne); got != want {
			t.Errorf("celebrate=%v: %d +1/+1 counters on Ash, want %d", celebrate, got, want)
		}
	}
}

func TestAshPartyCrasherHasHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ash := effectstest.PushCatalogPermanent(g, me.ID, "Ash, Party Crasher", "Legendary Creature — Human Peasant", effectstest.AshPartyCrasherOracle, true)
	c, ok := effectstest.CardOnBattlefield(g, ash)
	if !ok || !game.HasKeyword(&c, "haste") {
		t.Fatal("Ash lacks haste")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ash, g.Seats[1].ID); err != nil {
		t.Errorf("a summoning-sick Ash could not attack: %v", err)
	}
}

// --- Belligerent of the Ball -------------------------------------------------------

func TestBelligerentOfTheBallPumpsAndGrantsMenaceWhileCelebrating(t *testing.T) {
	// No celebration: nothing triggers as combat begins.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Belligerent of the Ball", "Creature — Ogre Warrior", effectstest.BelligerentOfTheBallOracle, false)
	effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStackFrom(g, "Belligerent of the Ball") {
		t.Fatal("the trigger fired with no celebration")
	}

	// Celebration: the chosen creature gets +1/+0 and menace until end of turn.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Belligerent of the Ball", "Creature — Ogre Warrior", effectstest.BelligerentOfTheBallOracle, false)
	bear := effectstest.B12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.CelebrationTokens(t, g2, me2.ID, 2)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PickCard(t, g2, me2.ID, bear)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CurrentPowerFor(t, g2, bear); got != 3 {
		t.Errorf("Bear power %d, want 3", got)
	}
	if c, _ := effectstest.CardOnBattlefield(g2, bear); !game.HasKeyword(&c, "menace") {
		t.Error("the Bear did not gain menace")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g2)
	if got := effectstest.CurrentPowerFor(t, g2, bear); got != 2 {
		t.Errorf("Bear power %d on the next turn, want 2", got)
	}
	if c, _ := effectstest.CardOnBattlefield(g2, bear); game.HasKeyword(&c, "menace") {
		t.Error("menace outlasted the turn")
	}
}

// --- Bespoke Battlegarb ------------------------------------------------------------

func TestBespokeBattlegarbAttachesAtCombatWhileCelebrating(t *testing.T) {
	// No celebration: the Battlegarb stays unattached.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	garb := effectstest.PushCatalogPermanent(g, me.ID, "Bespoke Battlegarb", "Artifact — Equipment", effectstest.BespokeBattlegarbOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.CardOnBattlefield(g, garb); c.AttachedTo.ID == bear {
		t.Fatal("attached with no celebration")
	}

	// The Battlegarb entering counts: it and one token are celebration.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	garb2 := effectstest.CelebrationEnterFromHand(t, g2, me2.ID, "Bespoke Battlegarb", "Artifact — Equipment", effectstest.BespokeBattlegarbOracle)
	bear2 := effectstest.B12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.CelebrationTokens(t, g2, me2.ID, 1)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PickCard(t, g2, me2.ID, bear2)
	effectstest.PassPriorityAroundTable(t, g2)
	c, ok := effectstest.CardOnBattlefield(g2, garb2)
	if !ok || c.AttachedTo.ID != bear2 {
		t.Fatalf("celebration: equipped to %v, want the Bear", c.AttachedTo.ID)
	}
	if got := effectstest.CurrentPowerFor(t, g2, bear2); got != 4 {
		t.Errorf("equipped Bear power %d, want 4 (+2/+0)", got)
	}
}

func TestBespokeBattlegarbHasEquipTwo(t *testing.T) {
	abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.BespokeBattlegarbOracle})
	if len(abs) != 1 || abs[0].Label != "Equip {2}" {
		t.Fatalf("activated abilities %+v, want one equip {2}", abs)
	}
}
