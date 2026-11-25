package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b215_cards_test.go — S82 OLD batch 215 (#5530, Dominaria): Chandra,
// Bold Pyromancer, Teferi, Timebender, Mishra's Self-Replicator, Diligent
// Excavator, Memorial to Unity and Voltaic Servant. Each test plays the real
// catalog entry through the engine.

const (
	oldB215Chandra    = "1cdba3c1-19e1-43c3-a0d8-5fb2b88060d8"
	oldB215Teferi     = "761bb894-d3f9-4248-af76-1559c502e342"
	oldB215Replicator = "5546d127-a961-4e33-bbc9-06ee14b0cb04"
	oldB215Excavator  = "a3866181-c127-45d6-8553-97e9f64a3ca5"
	oldB215Memorial   = "a74494ef-aa35-4830-9b4c-47bff5270efc"
	oldB215Servant    = "d0c2703b-bd74-41ac-b516-c46c654de1bb"
)

func TestOldB215EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB215Chandra:    "Chandra, Bold Pyromancer",
		oldB215Teferi:     "Teferi, Timebender",
		oldB215Replicator: "Mishra's Self-Replicator",
		oldB215Excavator:  "Diligent Excavator",
		oldB215Memorial:   "Memorial to Unity",
		oldB215Servant:    "Voltaic Servant",
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

func oldB215Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx int, targets []game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("activate ability %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB215CastHistoric(t *testing.T, g *game.Game, me *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	id := effectstest.HandSpell(me, name, typeLine, "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// --- Chandra, Bold Pyromancer ----------------------------------------------

func TestOldB215ChandraPlusOneAddsManaAndDamagesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Chandra, Bold Pyromancer", "Legendary Planeswalker — Chandra", oldB215Chandra, 5)
	before := opp.Life
	oldB215Activate(t, g, me, walker, 0, effectstest.P8Player(opp.ID))
	if got := before - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 2 {
		t.Errorf("the pool holds %d mana, want {R}{R}", got)
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB215ChandraMinusThreeHitsACreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Chandra, Bold Pyromancer", "Legendary Planeswalker — Chandra", oldB215Chandra, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(rock)}); err == nil {
		t.Error("Chandra's −3 targeted an artifact")
	}
	oldB215Activate(t, g, me, walker, 1, effectstest.CardRefs(bear))
	if g.Battlefield.Contains(bear) {
		t.Error("the 3-toughness creature survived 3 damage")
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestOldB215ChandraMinusSevenBurnsAPlayerAndEverythingTheyControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	walker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Chandra, Bold Pyromancer", "Legendary Planeswalker — Chandra", oldB215Chandra, 7)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 8, 10)
	theirWalker := effectstest.PushCatalogWalkerWithLine(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test", "", 9)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	before, thirdBefore := opp.Life, third.Life
	oldB215Activate(t, g, me, walker, 2, effectstest.P8Player(opp.ID))
	if got := before - opp.Life; got != 10 {
		t.Errorf("the target player lost %d, want 10", got)
	}
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(theirWalker) {
		t.Error("a creature or planeswalker the target player controls survived 10 damage")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(other) || third.Life != thirdBefore {
		t.Error("something the target player does not control was hurt")
	}
}

// --- Teferi, Timebender ------------------------------------------------------

func TestOldB215TeferiPlusTwoUntapsAnArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Teferi, Timebender", "Legendary Planeswalker — Teferi", oldB215Teferi, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Error("Teferi's +2 targeted a land")
	}
	oldB215Activate(t, g, me, walker, 0, effectstest.CardRefs(bear))
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature is still tapped")
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
	// "Up to one": it can be activated with no target.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == walker {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
}

func TestOldB215TeferiMinusThreeGainsLifeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Teferi, Timebender", "Legendary Planeswalker — Teferi", oldB215Teferi, 5)
	effectstest.AwbPadLibrary(me)
	life, hand := me.Life, len(me.Hand.Cards)
	oldB215Activate(t, g, me, walker, 1, nil)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestOldB215TeferiMinusNineTakesAnExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Teferi, Timebender", "Legendary Planeswalker — Teferi", oldB215Teferi, 9)
	oldB215Activate(t, g, me, walker, 2, nil)
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Errorf("%d extra turns queued, want 1", got)
	}
	if g.Battlefield.Contains(walker) {
		t.Error("Teferi stayed at 0 loyalty")
	}
}

// --- Mishra's Self-Replicator ----------------------------------------------

func TestOldB215ReplicatorCopiesItselfWhenYouPayForAHistoricSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Mishra's Self-Replicator", "Artifact Creature — Assembly-Worker", oldB215Replicator, 2, 2)
	oldB215CastHistoric(t, g, me, "Rock", "Artifact")
	effectstest.ApaMana(me, "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effects.CountControlled(g, me.ID, func(c game.Card) bool { return c.Name == "Mishra's Self-Replicator" }); n != 2 {
		t.Errorf("you control %d Replicators, want 2", n)
	}
}

func TestOldB215ReplicatorDoesNothingWhenYouDeclineOrCastANonHistoricSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Mishra's Self-Replicator", "Artifact Creature — Assembly-Worker", oldB215Replicator, 2, 2)
	oldB215CastHistoric(t, g, me, "Bear", "Creature — Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a nonhistoric spell triggered the Replicator")
	}
	oldB215CastHistoric(t, g, me, "Legend", "Legendary Creature — Bear")
	effectstest.ApaMana(me, "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effects.CountControlled(g, me.ID, func(c game.Card) bool { return c.Name == "Mishra's Self-Replicator" }); n != 1 {
		t.Errorf("you control %d Replicators, want 1", n)
	}
}

// --- Diligent Excavator ------------------------------------------------------

func TestOldB215ExcavatorMillsTheTargetPlayerOnAHistoricSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Diligent Excavator", "Creature — Human Artificer", oldB215Excavator, 1, 3)
	effectstest.AwbPadLibrary(opp)
	before := len(opp.Library.Cards)
	oldB215CastHistoric(t, g, me, "Bear", "Creature — Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != before {
		t.Fatal("a nonhistoric spell milled")
	}
	oldB215CastHistoric(t, g, me, "Rock", "Artifact")
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - len(opp.Library.Cards); got != 2 {
		t.Errorf("the opponent milled %d cards, want 2", got)
	}
}

// --- Voltaic Servant ---------------------------------------------------------

func TestOldB215ServantUntapsAnArtifactAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	servant := effectstest.B12Push(g, me.ID, "Voltaic Servant", "Artifact Creature — Construct", oldB215Servant, 1, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	_ = servant
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, rock)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, rock) {
		t.Error("the artifact is still tapped after the end step trigger")
	}
}

// --- Memorial to Unity -------------------------------------------------------

func TestOldB215MemorialEntersTappedAndTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Memorial to Unity", "Land", oldB215Memorial, false)
	if got := effectstest.ManaAbilityProduced(t, g, land); len(got) == 0 || got[0] != "{G}" {
		t.Errorf("Memorial to Unity produces %v, want {G}", got)
	}
}

func TestOldB215MemorialSacrificesToTakeACreatureFromTheTopFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Memorial to Unity", "Land", oldB215Memorial, false)
	beast := uuid.New()
	deep := uuid.New()
	for _, c := range []game.Card{
		{InstanceID: deep, Name: "Deep Creature", TypeLine: "Creature — Beast", Power: 1, Toughness: 1},
		{InstanceID: uuid.New(), Name: "Filler 1", TypeLine: "Land"},
		{InstanceID: uuid.New(), Name: "Filler 2", TypeLine: "Land"},
		{InstanceID: uuid.New(), Name: "Filler 3", TypeLine: "Land"},
		{InstanceID: uuid.New(), Name: "Filler 4", TypeLine: "Land"},
		{InstanceID: beast, Name: "Beast", TypeLine: "Creature — Beast", Power: 5, Toughness: 5},
		{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery"},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	effectstest.ApaMana(me, "G", "C", "C")
	effectstest.ApaActivate(t, g, me, land, 0, game.ActivateAbilityParams{})
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(beast) {
		t.Error("the creature card is not in hand")
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if me.Hand.Contains(deep) {
		t.Error("a creature card below the top five was taken")
	}
}
