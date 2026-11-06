package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_celebration_cards_test.go — celebration (CR 207.2c, #5164) and the
// Wilds of Eldraine cards that read it. The seam itself (the tally cell)
// is pinned in game/turn_tally_celebration_test.go; these play each
// card's real catalog entry.

const (
	armoryMiceOracle            = "68f7d1bf-48c7-483e-b17f-a60008854b46"
	ashPartyCrasherOracle       = "1bc701e2-913a-45d5-bc3f-16f87e7e2653"
	belligerentOfTheBallOracle  = "2be2f6c8-d8df-4e5c-a0c1-7566183bbe81"
	bespokeBattlegarbOracle     = "b2ff15b8-1be4-4f94-b4cc-aa959a50288e"
	celebrationSoldierTokenName = "Soldier"
)

func TestCelebrationCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		armoryMiceOracle:           "Armory Mice",
		ashPartyCrasherOracle:      "Ash, Party Crasher",
		belligerentOfTheBallOracle: "Belligerent of the Ball",
		bespokeBattlegarbOracle:    "Bespoke Battlegarb",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// celebrationTokens creates n 1/1 Soldier tokens under owner's control:
// n nonland permanents entering.
func celebrationTokens(t *testing.T, g *game.Game, owner uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, game.Card{Name: celebrationSoldierTokenName, TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}, n); err != nil {
			t.Fatalf("create tokens: %v", err)
		}
	})
}

// celebrationEnterFromHand puts a card from owner's hand onto the
// battlefield through the ordinary entry pipeline, so its entry counts.
func celebrationEnterFromHand(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		g.PlayerByIDForEffect(owner).Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: 1, Toughness: 1, Owner: owner, Controller: owner,
		})
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: owner})
	})
	if err != nil || entered == uuid.Nil {
		t.Fatalf("put %s onto the battlefield: %v", name, err)
	}
	return entered
}

// celebrationCountFor is the seam's count, read under the lock.
func celebrationCountFor(g *game.Game, you uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.NonlandPermanentsEnteredThisTurn(you) })
	return n
}

func celebrationToughnessFor(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := cardOnBattlefield(g, id)
	if !ok {
		t.Fatalf("card %s not on battlefield", id)
	}
	return c.CurrentToughness()
}

// --- the condition helper ---------------------------------------------------

// Celebration is two or more: one entry is not enough, a land does not
// help, and an opponent's entries are theirs.
func TestCelebrationNeedsTwoNonlandPermanentsOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	holds := func() bool {
		var ok bool
		g.WithWriteLock(func() { ok = Celebration(g, me.ID) })
		return ok
	}
	if holds() {
		t.Fatal("celebration before anything entered")
	}
	celebrationTokens(t, g, me.ID, 1)
	celebrationEnterFromHand(t, g, me.ID, "Forest", "Basic Land — Forest", "")
	celebrationTokens(t, g, opp.ID, 3)
	if holds() {
		t.Errorf("celebration with one token, a land and an opponent's tokens (count %d)", celebrationCountFor(g, me.ID))
	}
	celebrationEnterFromHand(t, g, me.ID, "Ornithopter", "Artifact Creature — Thopter", "")
	if !holds() {
		t.Errorf("no celebration after a token and an artifact creature (count %d)", celebrationCountFor(g, me.ID))
	}
}

// --- Armory Mice ---------------------------------------------------------------

func TestArmoryMiceGetsPlusZeroPlusTwoWhileCelebrating(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mice := pushCatalogPermanent(g, me.ID, "Armory Mice", "Creature — Mouse", armoryMiceOracle, false)
	base := celebrationCountFor(g, me.ID)
	if base >= 2 {
		t.Fatalf("setup: already %d entries", base)
	}
	if got := celebrationToughnessFor(t, g, mice); got != 1 {
		t.Fatalf("toughness %d with no celebration, want 1", got)
	}
	celebrationTokens(t, g, me.ID, 2-base)
	if got := celebrationToughnessFor(t, g, mice); got != 3 {
		t.Errorf("toughness %d while celebrating, want 3", got)
	}
	if got := currentPowerFor(t, g, mice); got != 1 {
		t.Errorf("power %d while celebrating, want 1 (+0/+2)", got)
	}
	// The tokens leaving does not end it: they entered this turn.
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == celebrationSoldierTokenName {
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
	advanceToNextSeatsTurn(t, g)
	if got := celebrationToughnessFor(t, g, mice); got != 1 {
		t.Errorf("toughness %d on the next turn, want 1", got)
	}
}

// The Mice count themselves: the Mice and one other nonland permanent
// entering is celebration; the Mice and a land is not.
func TestArmoryMiceCountsItselfEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mice := celebrationEnterFromHand(t, g, me.ID, "Armory Mice", "Creature — Mouse", armoryMiceOracle)
	celebrationEnterFromHand(t, g, me.ID, "Forest", "Basic Land — Forest", "")
	if got := celebrationToughnessFor(t, g, mice); got != 1 {
		t.Errorf("the Mice and a land: toughness %d, want 1", got)
	}
	celebrationTokens(t, g, me.ID, 1)
	if got := celebrationToughnessFor(t, g, mice); got != 3 {
		t.Errorf("the Mice and a token: toughness %d, want 3", got)
	}
}

// --- Ash, Party Crasher ------------------------------------------------------------

func TestAshPartyCrasherGetsACounterWhenAttackingWhileCelebrating(t *testing.T) {
	for _, celebrate := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ash := pushCatalogPermanent(g, me.ID, "Ash, Party Crasher", "Legendary Creature — Human Peasant", ashPartyCrasherOracle, false)
		if celebrate {
			celebrationTokens(t, g, me.ID, 2)
		}
		declareAttack(t, g, opp.ID, ash)
		passPriorityAroundTable(t, g)
		want := 0
		if celebrate {
			want = 1
		}
		if got := counterCount(g, ash, game.CounterPlusOne); got != want {
			t.Errorf("celebrate=%v: %d +1/+1 counters on Ash, want %d", celebrate, got, want)
		}
	}
}

func TestAshPartyCrasherHasHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ash := pushCatalogPermanent(g, me.ID, "Ash, Party Crasher", "Legendary Creature — Human Peasant", ashPartyCrasherOracle, true)
	c, ok := cardOnBattlefield(g, ash)
	if !ok || !game.HasKeyword(&c, "haste") {
		t.Fatal("Ash lacks haste")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ash, g.Seats[1].ID); err != nil {
		t.Errorf("a summoning-sick Ash could not attack: %v", err)
	}
}

// --- Belligerent of the Ball -------------------------------------------------------

func TestBelligerentOfTheBallPumpsAndGrantsMenaceWhileCelebrating(t *testing.T) {
	// No celebration: nothing triggers as combat begins.
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Belligerent of the Ball", "Creature — Ogre Warrior", belligerentOfTheBallOracle, false)
	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if latestPickTarget(g, me.ID) != nil || triggerOnStackFrom(g, "Belligerent of the Ball") {
		t.Fatal("the trigger fired with no celebration")
	}

	// Celebration: the chosen creature gets +1/+0 and menace until end of turn.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Belligerent of the Ball", "Creature — Ogre Warrior", belligerentOfTheBallOracle, false)
	bear := b12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	celebrationTokens(t, g2, me2.ID, 2)
	advanceTo(t, g2, game.StepBeginCombat)
	pickCard(t, g2, me2.ID, bear)
	passPriorityAroundTable(t, g2)
	if got := currentPowerFor(t, g2, bear); got != 3 {
		t.Errorf("Bear power %d, want 3", got)
	}
	if c, _ := cardOnBattlefield(g2, bear); !game.HasKeyword(&c, "menace") {
		t.Error("the Bear did not gain menace")
	}
	advanceToNextSeatsTurn(t, g2)
	if got := currentPowerFor(t, g2, bear); got != 2 {
		t.Errorf("Bear power %d on the next turn, want 2", got)
	}
	if c, _ := cardOnBattlefield(g2, bear); game.HasKeyword(&c, "menace") {
		t.Error("menace outlasted the turn")
	}
}

// --- Bespoke Battlegarb ------------------------------------------------------------

func TestBespokeBattlegarbAttachesAtCombatWhileCelebrating(t *testing.T) {
	// No celebration: the Battlegarb stays unattached.
	g := newCatalogGame(t)
	me := g.Seats[0]
	garb := pushCatalogPermanent(g, me.ID, "Bespoke Battlegarb", "Artifact — Equipment", bespokeBattlegarbOracle, false)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c, _ := cardOnBattlefield(g, garb); c.AttachedTo.ID == bear {
		t.Fatal("attached with no celebration")
	}

	// The Battlegarb entering counts: it and one token are celebration.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	garb2 := celebrationEnterFromHand(t, g2, me2.ID, "Bespoke Battlegarb", "Artifact — Equipment", bespokeBattlegarbOracle)
	bear2 := b12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	celebrationTokens(t, g2, me2.ID, 1)
	advanceTo(t, g2, game.StepBeginCombat)
	pickCard(t, g2, me2.ID, bear2)
	passPriorityAroundTable(t, g2)
	c, ok := cardOnBattlefield(g2, garb2)
	if !ok || c.AttachedTo.ID != bear2 {
		t.Fatalf("celebration: equipped to %v, want the Bear", c.AttachedTo.ID)
	}
	if got := currentPowerFor(t, g2, bear2); got != 4 {
		t.Errorf("equipped Bear power %d, want 4 (+2/+0)", got)
	}
}

func TestBespokeBattlegarbHasEquipTwo(t *testing.T) {
	abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: bespokeBattlegarbOracle})
	if len(abs) != 1 || abs[0].Label != "Equip {2}" {
		t.Fatalf("activated abilities %+v, want one equip {2}", abs)
	}
}
