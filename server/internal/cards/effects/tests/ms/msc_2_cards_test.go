package testsms

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_2_cards_test.go — Marvel Super Heroes Commander, second batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc2GallantOracle    = "2fcb1c1c-6157-4955-b899-e8a8b337d188"
	msc2GrapplersOracle  = "9f2e03fb-45c0-442d-a716-a472373c72d0"
	msc2MassifOracle     = "23732d20-aba6-46cc-89d0-0542629bc6b5"
	msc2IntrepidOracle   = "cb8889c8-8ee3-4e27-9f47-ad1021ff61bb"
	msc2MoveOracle       = "8226f31d-6f51-49c3-87f7-0c68f7f4f9ce"
	msc2MandroidOracle   = "95ab1529-0de0-416a-82b4-c5ed9276a9b4"
	msc2HeraldOracle     = "74255add-2894-4c31-81c1-d38873fa325f"
	msc2PowerBoostOracle = "fed58ccc-db56-4094-8b56-bd29b00e9327"
)

func TestMsc2EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc2GallantOracle:    "Gallant Citizen",
		msc2GrapplersOracle:  "Glamorous Grapplers",
		msc2MassifOracle:     "Glittering Massif",
		msc2IntrepidOracle:   "Intrepid Ace",
		msc2MoveOracle:       "Make Your Move",
		msc2MandroidOracle:   "Mandroid Squadron",
		msc2HeraldOracle:     "Mist-Cloaked Herald",
		msc2PowerBoostOracle: "Power Boost",
	}
	for oracle, name := range want {
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

func TestMsc2GallantCitizenDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	lib := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Gallant Citizen", "Creature — Human Citizen", msc2GallantOracle, 1, 1)
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("library fell by %d, want 1", got)
	}
}

func TestMsc2GlamorousGrapplersHasMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grapplers := effectstest.B12Push(g, me.ID, "Glamorous Grapplers", "Creature — Human Performer Villain", msc2GrapplersOracle, 3, 3)
	effectstest.AssertKeywords(t, g, grapplers, "menace")
}

func TestMsc2GlitteringMassifEntersTappedTapsForEitherColourAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Glittering Massif", msc2MassifOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Glittering Massif")
	for _, colour := range []string{"R", "W"} {
		me.ManaPool = nil
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		effectstest.B28TapForMana(t, g, me.ID, id, colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	cycled, _ := effectstest.CycleFromHand(t, g2, "Glittering Massif", "Land — Mountain Plains", msc2MassifOracle, "{C}{C}")
	if !me2.Graveyard.Contains(cycled) {
		t.Error("the cycled land is not in the graveyard")
	}
}

func TestMsc2IntrepidAceGetsPlusTwoOnlyWhenNotAttackingOrBlocking(t *testing.T) {
	// Idle: the +2/+0 applies.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	idle := effectstest.B12Push(g, me.ID, "Intrepid Ace", "Creature — Human Pilot", msc2IntrepidOracle, 2, 1)
	if p, tg := effectstest.EffectivePower(t, g, idle), effectstest.EffectiveToughness(t, g, idle); p != 4 || tg != 1 {
		t.Errorf("idle Ace is %d/%d, want 4/1", p, tg)
	}

	// Attacking: the bonus goes away.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	attacking := effectstest.B12Push(g2, me2.ID, "Intrepid Ace", "Creature — Human Pilot", msc2IntrepidOracle, 2, 1)
	g2.WithWriteLock(func() {
		for i := range g2.Battlefield.Cards {
			if g2.Battlefield.Cards[i].InstanceID == attacking {
				g2.Battlefield.Cards[i].AttackingTarget = opp2.ID
			}
		}
	})
	if p := effectstest.EffectivePower(t, g2, attacking); p != 2 {
		t.Errorf("attacking Ace has power %d, want 2", p)
	}

	// Blocking: the bonus goes away too.
	g3 := effectstest.NewCatalogGame(t)
	me3, opp3 := g3.Seats[0], g3.Seats[1]
	blocker := effectstest.B12Push(g3, me3.ID, "Intrepid Ace", "Creature — Human Pilot", msc2IntrepidOracle, 2, 1)
	attacker := effectstest.B12Creature(g3, opp3.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CrtBlock(g3, blocker, attacker, me3.ID)
	if p := effectstest.EffectivePower(t, g3, blocker); p != 2 {
		t.Errorf("blocking Ace has power %d, want 2", p)
	}
}

func TestMsc2MakeYourMoveDestroysArtifactEnchantmentOrBigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Four Power", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Make Your Move", "Instant", msc2MoveOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a creature with power 4 survived")
	}

	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	relic := effectstest.B12Push(g2, opp2.ID, "Relic", "Artifact", "", 0, 0)
	effectstest.CastCatalogSpell(t, g2, "Make Your Move", "Instant", msc2MoveOracle, effectstest.B16TargetCard(relic))
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(relic) {
		t.Error("an artifact survived")
	}

	g3 := effectstest.NewCatalogGame(t)
	opp3 := g3.Seats[1]
	aura := effectstest.B12Push(g3, opp3.ID, "Pact", "Enchantment", "", 0, 0)
	effectstest.CastCatalogSpell(t, g3, "Make Your Move", "Instant", msc2MoveOracle, effectstest.B16TargetCard(aura))
	effectstest.PassPriorityAroundTable(t, g3)
	if g3.Battlefield.Contains(aura) {
		t.Error("an enchantment survived")
	}
}

func TestMsc2MakeYourMoveRejectsASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Three Power", "Creature — Bear", 3, 3)
	err := effectstest.CastCatalogSpellErr(t, g, "Make Your Move", "Instant", msc2MoveOracle, effectstest.B16TargetCard(small))
	if err == nil {
		t.Error("a creature with power 3 was accepted as the target")
	}
}

func TestMsc2MandroidSquadronGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Mandroid Squadron", "Artifact Creature — Human Soldier", msc2MandroidOracle, 0, 4)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestMsc2MistCloakedHeraldCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herald := effectstest.B12Push(g, me.ID, "Mist-Cloaked Herald", "Creature — Merfolk Warrior", msc2HeraldOracle, 1, 1)
	effectstest.AssertRestrictions(t, g, herald, game.CantBeBlocked)
}

func TestMsc2PowerBoostAddsOnePowerToYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Power Boost", "Enchantment", msc2PowerBoostOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tg := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tg != 2 {
		t.Errorf("my bear is %d/%d, want 3/2", p, tg)
	}
	if p, tg := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("their bear is %d/%d, want 2/2", p, tg)
	}
}
