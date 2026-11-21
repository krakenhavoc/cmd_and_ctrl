package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Slobad, Iron Goblin --------------------------------------------------------

func s82OneB18Artifact(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", ManaCost: cost,
		Owner: owner, Controller: owner,
	})
}

func TestS82OneB18SlobadAddsRedEqualToTheSacrificedArtifactsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slobad := effectstest.PushCatalogPermanent(g, me.ID, "Slobad, Iron Goblin", "Legendary Creature — Phyrexian Goblin Artificer", s82OneB18SlobadOracle, false)
	relic := s82OneB18Artifact(g, me.ID, "Costly Relic", "{3}")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, slobad, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("R"); got != 3 {
		t.Errorf("pool holds %d {R}, want 3 (the sacrificed artifact's mana value)", got)
	}
	if effects.OnBattlefield(g, relic) {
		t.Error("the artifact was not sacrificed")
	}
	if !effectstest.LayeredCard(t, g, slobad).Tapped {
		t.Error("Slobad did not tap")
	}
}

func TestS82OneB18SlobadsManaPaysForArtifactsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slobad := effectstest.PushCatalogPermanent(g, me.ID, "Slobad, Iron Goblin", "Legendary Creature — Phyrexian Goblin Artificer", s82OneB18SlobadOracle, false)
	relic := s82OneB18Artifact(g, me.ID, "Costly Relic", "{3}")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, slobad, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, Strict: true}); err == nil {
		t.Error("the restricted mana paid for a non-artifact spell")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Cheap Trinket", "Artifact", "{R}{R}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the restricted mana could not pay for an artifact spell: %v", err)
	}
}

func TestS82OneB18SlobadSacrificingAZeroValueArtifactAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slobad := effectstest.PushCatalogPermanent(g, me.ID, "Slobad, Iron Goblin", "Legendary Creature — Phyrexian Goblin Artificer", s82OneB18SlobadOracle, false)
	token := s82OneB18Artifact(g, me.ID, "Zero Relic", "")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, slobad, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{token}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool holds %d mana, want 0", got)
	}
}

func TestS82OneB18SlobadNeedsAnArtifactToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slobad := effectstest.PushCatalogPermanent(g, me.ID, "Slobad, Iron Goblin", "Legendary Creature — Phyrexian Goblin Artificer", s82OneB18SlobadOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, slobad, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("sacrificed a non-artifact")
	}
	if err := g.ActivateManaAbility(me.ID, slobad, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("activated with nothing sacrificed")
	}
}

// --- Melira, the Living Cure ----------------------------------------------------

func TestS82OneB18MeliraCapsPoisonAtOneThenNoneMoreThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Melira, the Living Cure", "Legendary Creature — Human Scout", s82OneB18MeliraOracle, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 3); err != nil {
			t.Fatal(err)
		}
	})
	if got := me.Counters[game.CounterPoison]; got != 1 {
		t.Fatalf("poison %d after three would-be counters, want 1", got)
	}
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := me.Counters[game.CounterPoison]; got != 1 {
		t.Errorf("poison %d after more, want it to stay 1 this turn", got)
	}
	// Opponents are unaffected.
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 3); err != nil {
			t.Fatal(err)
		}
	})
	if got := opp.Counters[game.CounterPoison]; got != 3 {
		t.Errorf("an opponent got %d poison counters, want 3", got)
	}
}

func TestS82OneB18MeliraTheBanLastsOnlyThisTurnAndOutlivesMelira(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	melira := effectstest.PushCatalogPermanent(g, me.ID, "Melira, the Living Cure", "Legendary Creature — Human Scout", s82OneB18MeliraOracle, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 2) })
	effectstest.B27Kill(g, melira) // Melira leaves; the "can't get additional" effect remains this turn.
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 2) })
	if got := me.Counters[game.CounterPoison]; got != 1 {
		t.Errorf("poison %d, want 1 (the ban outlives Melira)", got)
	}
	// Next turn, with no Melira, the counters arrive in full.
	effectstest.AdvanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 2) })
	if got := me.Counters[game.CounterPoison]; got != 3 {
		t.Errorf("poison %d on a later turn without Melira, want 3", got)
	}
}

func TestS82OneB18MeliraExileReturnsTheCreatureThatDiesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	melira := effectstest.PushCatalogPermanent(g, me.ID, "Melira, the Living Cure", "Legendary Creature — Human Scout", s82OneB18MeliraOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// Melira herself is not "another".
	if err := g.ActivateCatalogAbility(me.ID, melira, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(melira), Strict: true}); err == nil {
		t.Fatal("Melira chose herself")
	}
	if err := g.ActivateCatalogAbility(me.ID, melira, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(melira); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("Melira is in %v, want exile", z)
	}
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Fatal("the creature did not return")
	}
	if c := effectstest.LayeredCard(t, g, bear); c.Controller != opp.ID {
		t.Errorf("returned under %v, want its owner %v", c.Controller, opp.ID)
	}
}

func TestS82OneB18MeliraExileAlsoWorksOnAnArtifactAndLapsesAfterTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	melira := effectstest.PushCatalogPermanent(g, me.ID, "Melira, the Living Cure", "Legendary Creature — Human Scout", s82OneB18MeliraOracle, false)
	relic := s82OneB18Artifact(g, me.ID, "Fragile Relic", "{2}")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, melira, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(relic), Strict: true}); err != nil {
		t.Fatalf("activate on an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The turn ends; the next turn it dies and stays dead.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B27Kill(g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, relic) {
		t.Error("the artifact returned from a graveyard on a later turn")
	}
}

// --- Encroaching Mycosynth ------------------------------------------------------

func TestS82OneB18MycosynthMakesYourNonlandPermanentsSpellsAndCardsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
	land := effectstest.B12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myGrave := effectstest.RedCard(me.ID, "Graveyard Bear", "Creature — Bear")
	myGraveLand := effectstest.RedCard(me.ID, "Graveyard Land", "Land")
	mySorcery := effectstest.RedCard(me.ID, "My Sorcery", "Sorcery")
	theirGrave := effectstest.RedCard(opp.ID, "Their Graveyard Bear", "Creature — Bear")
	me.Graveyard.PushTop(myGrave)
	me.Graveyard.PushTop(myGraveLand)
	me.Hand.PushTop(mySorcery)
	opp.Graveyard.PushTop(theirGrave)
	effectstest.PushCatalogPermanent(g, me.ID, "Encroaching Mycosynth", "Artifact", s82OneB18MycosynthOracle, false)
	spell := effectstest.CastCatalogSpell(t, g, "Cast Bear", "Creature — Bear", "", nil)
	instant := effectstest.CastCatalogSpell(t, g, "Cast Shock", "Instant", "", nil)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })

	for _, tc := range []struct {
		name string
		c    game.Card
		want bool
	}{
		{"my creature", effectstest.ZoneLayerCard(t, g, mine), true},
		{"my enchantment", effectstest.ZoneLayerCard(t, g, ench), true},
		{"my land", effectstest.ZoneLayerCard(t, g, land), false},
		{"their creature", effectstest.ZoneLayerCard(t, g, theirs), false},
		{"my graveyard creature card", effectstest.ZoneLayerCard(t, g, myGrave.InstanceID), true},
		{"my graveyard land card", effectstest.ZoneLayerCard(t, g, myGraveLand.InstanceID), false},
		{"my sorcery card", effectstest.ZoneLayerCard(t, g, mySorcery.InstanceID), false},
		{"their graveyard creature card", effectstest.ZoneLayerCard(t, g, theirGrave.InstanceID), false},
		{"my permanent spell", effectstest.ZoneLayerCard(t, g, spell), true},
		{"my instant spell", effectstest.ZoneLayerCard(t, g, instant), false},
	} {
		if got := tc.c.IsArtifact(); got != tc.want {
			t.Errorf("%s: is an artifact = %v, want %v", tc.name, got, tc.want)
		}
		if tc.want && tc.c.TypeLine == "" {
			t.Errorf("%s lost its printed types", tc.name)
		}
	}
	if c := effectstest.ZoneLayerCard(t, g, mine); !c.IsCreature() {
		t.Error("my creature stopped being a creature")
	}
}

func TestS82OneB18MycosynthEndsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	myr := effectstest.PushCatalogPermanent(g, me.ID, "Encroaching Mycosynth", "Artifact", s82OneB18MycosynthOracle, false)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if !effectstest.ZoneLayerCard(t, g, mine).IsArtifact() {
		t.Fatal("setup: not an artifact under the Mycosynth")
	}
	effectstest.B27Kill(g, myr)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if effectstest.ZoneLayerCard(t, g, mine).IsArtifact() {
		t.Error("still an artifact after the Mycosynth left")
	}
}
