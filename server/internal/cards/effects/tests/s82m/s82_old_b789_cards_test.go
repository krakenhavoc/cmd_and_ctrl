package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b789_cards_test.go — Visions, S82 OLD batch 789 (#5631). Every test
// plays the real catalog entry through the engine.

const (
	oldB789Gorilla  = "1dd6df3f-eac4-4f30-b832-4a3f3a97632f"
	oldB789Funeral  = "8866272d-89cb-478c-82ae-20e1d510eab7"
	oldB789Song     = "ab4b6524-3667-4db3-b5cb-1b02ee944906"
	oldB789Shear    = "fa683059-b9de-4063-ae71-597ed4e2df30"
	oldB789Mask     = "fce29dd8-23b1-465c-a977-526f716dc8f6"
	oldB789Corrode  = "3d3fd8a2-78d4-4f52-8fcd-0c6fa93f0043"
	oldB789Hellkite = "533e41cc-7cae-47f2-886b-8a3ffb82bf81"
	oldB789Grass    = "53cc4b93-538d-4193-8bbb-5b73d3ce9d79"
	oldB789Bubble   = "9344ae2e-cb2e-400d-9157-e8158ce31813"
	oldB789Wave     = "d8f08699-6210-41c3-bfa0-c4b3bcbbaf1c"
)

func TestOldB789EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB789Gorilla: "Raging Gorilla", oldB789Funeral: "Funeral Charm", oldB789Song: "Song of Blood",
		oldB789Shear: "Wind Shear", oldB789Mask: "Dragon Mask", oldB789Corrode: "Corrosion",
		oldB789Hellkite: "Firestorm Hellkite", oldB789Grass: "Elephant Grass", oldB789Bubble: "Juju Bubble",
		oldB789Wave: "Heat Wave",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB789Upkeep walks to seat 1's upkeep with the permanent carrying `seed` age
// counters and returns the open payment prompt (the upkeep adds one more).
func oldB789Upkeep(t *testing.T, g *game.Game, id uuid.UUID, seed int) *game.PendingChoice {
	t.Helper()
	if seed > 0 {
		effectstest.AddCountersForTest(g, id, game.CounterAge, seed)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.RemoraUpkeepPrompt(t, g, g.Seats[1].ID)
}

func TestOldB789RagingGorillaShrinksOncePerBlockOrBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorilla := effectstest.B12Push(g, me.ID, "Raging Gorilla", "Creature — Ape", oldB789Gorilla, 2, 3)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: gorilla, Target: a, Amount: 1})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, gorilla); p != 4 || tough != 1 {
		t.Fatalf("after blocking the Gorilla is %d/%d, want 4/1", p, tough)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	gorilla = effectstest.B12Push(g, me.ID, "Raging Gorilla", "Creature — Ape", oldB789Gorilla, 2, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: gorilla, CardID: gorilla, Target: gorilla})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, gorilla); p != 4 || tough != 1 {
		t.Errorf("after becoming blocked the Gorilla is %d/%d, want 4/1", p, tough)
	}
}

func TestOldB789FirestormHellkiteFliesTramplesAndChargesUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	kite := effectstest.B12Push(g, me.ID, "Firestorm Hellkite", "Creature — Dragon", oldB789Hellkite, 6, 6)
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.P1g2HasKeyword(g, kite, kw) {
			t.Errorf("the Hellkite has no %s", kw)
		}
	}
	prompt := oldB789Upkeep(t, g, kite, 1)
	if prompt.PayCost != "{U}{R}{U}{R}" && prompt.PayCost != "{U}{U}{R}{R}" {
		t.Errorf("cost with two age counters = %q, want {U}{R} twice", prompt.PayCost)
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(kite) {
		t.Error("declining the upkeep did not sacrifice the Hellkite")
	}
}

func TestOldB789ElephantGrassBarsBlackAndTaxesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, opp.ID, "Elephant Grass", "Enchantment", oldB789Grass, false)
	black := effectstest.SeedColoredCreature(g, me.ID, "Black Bear", []string{"B"}, 0)
	green := effectstest.SeedColoredCreature(g, me.ID, "Green Bear", []string{"G"}, 0)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(black, opp.ID); err == nil {
		t.Fatal("a black creature attacked the Elephant Grass's controller")
	}
	if err := g.DeclareAttacker(black, other.ID); err != nil {
		t.Fatalf("a black creature could not attack another player: %v", err)
	}
	price := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: green, Target: opp.ID}})
	if price.Total.Generic != 2 {
		t.Errorf("a nonblack attacker owes %d, want {2} (%q)", price.Total.Generic, price.Cost)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: green, Target: other.ID}}); !p.IsFree() {
		t.Errorf("attacking a player without the Grass costs %q", p.Cost)
	}
	effectstest.FillPool(me, 2)
	if err := g.DeclareAttacker(green, opp.ID); err != nil {
		t.Errorf("a nonblack creature could not attack with {2} paid: %v", err)
	}
}

func TestOldB789HeatWaveBarsBlueAndChargesLifeToTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Heat Wave", "Enchantment", oldB789Wave, false)
	attacker := effectstest.SeedColoredCreature(g, me.ID, "Attacker", []string{"R"}, 0)
	blue := effectstest.SeedColoredCreature(g, opp.ID, "Blue Blocker", []string{"U"}, 0)
	red := effectstest.SeedColoredCreature(g, opp.ID, "Red Blocker", []string{"R"}, 0)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(blue, attacker); err == nil {
		t.Fatal("a blue creature blocked a creature the Heat Wave's controller controls")
	}
	if err := g.DeclareBlocker(red, attacker); err != nil {
		t.Fatalf("staging a nonblue block: %v", err)
	}
	if p := g.PriceBlockDeclaration(opp.ID); p.Life != 1 {
		t.Errorf("a nonblue blocker owes %d life, want 1", p.Life)
	}
	life := effectstest.LifeOf(g, opp.ID)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("finishing the block: %v", err)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-1 {
		t.Errorf("life %d → %d, want -1", life, got)
	}
}

func TestOldB789JujuBubbleGainsLifeAndGoesWhenYouPlayACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bubble := effectstest.PushCatalogPermanent(g, me.ID, "Juju Bubble", "Artifact", oldB789Bubble, false)
	effectstest.FillPool(me, 2)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.B16Activate(t, g, me.ID, bubble, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d → %d, want +1", life, got)
	}
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bubble) {
		t.Error("playing a land did not sacrifice the Juju Bubble")
	}
}

func TestOldB789JujuBubbleSurvivesAnOpponentsPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bubble := effectstest.PushCatalogPermanent(g, other.ID, "Juju Bubble", "Artifact", oldB789Bubble, false)
	_ = active
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bubble) {
		t.Error("an opponent playing a land sacrificed the Juju Bubble")
	}
}

func TestOldB789DragonMaskPumpsThenReturnsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mask := effectstest.PushCatalogPermanent(g, me.ID, "Dragon Mask", "Artifact", oldB789Mask, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 3)
	effectstest.B16Activate(t, g, me.ID, mask, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("pumped Bear is %d/%d, want 4/4", p, tough)
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not returned at the end step")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
}

func TestOldB789DragonMaskCantTargetAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mask := effectstest.PushCatalogPermanent(g, me.ID, "Dragon Mask", "Artifact", oldB789Mask, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, mask, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("Dragon Mask targeted a creature its controller does not control")
	}
}

func TestOldB789WindShearShrinksAndGroundsAttackingFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 3, Toughness: 3,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	ground := effectstest.B12Creature(g, me.ID, "Ground", "Creature — Bear", 3, 3)
	idleFlyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Idle Flyer", TypeLine: "Creature — Bird", Power: 3, Toughness: 3,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.BrAttack(t, g, flyer, ground)
	effectstest.CastInPlace(t, g, me.ID, "Wind Shear", oldB789Shear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, flyer); p != 1 || tough != 1 {
		t.Errorf("the attacking flyer is %d/%d, want 1/1", p, tough)
	}
	if effectstest.P1g2HasKeyword(g, flyer, "flying") {
		t.Error("the attacking flyer kept flying")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, ground); p != 3 || tough != 3 {
		t.Errorf("an attacker without flying is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, idleFlyer); p != 3 || tough != 3 || !effectstest.P1g2HasKeyword(g, idleFlyer, "flying") {
		t.Errorf("a non-attacking flyer changed: %d/%d", p, tough)
	}
}

func TestOldB789FuneralCharmEachMode(t *testing.T) {
	cast := func(t *testing.T, g *game.Game, mode int, ref game.TargetRef) {
		t.Helper()
		effectstest.AdvanceToMain(t, g)
		effectstest.P1g3CastModal(t, g, "Funeral Charm", "Instant", oldB789Funeral, []int{mode}, ref)
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Run("discard", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		before := opp.Hand.Size()
		cast(t, g, 0, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0})
		effectstest.DiscardFromHand(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if got := opp.Hand.Size(); got != before-1 {
			t.Errorf("hand %d → %d, want one fewer", before, got)
		}
	})
	t.Run("pump", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		cast(t, g, 1, effectstest.P1g3ModeCard(bear, 0))
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 1 {
			t.Errorf("Bear is %d/%d, want 4/1", p, tough)
		}
	})
	t.Run("swampwalk", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		cast(t, g, 2, effectstest.P1g3ModeCard(bear, 0))
		if !effectstest.P1g2HasKeyword(g, bear, "swampwalk") {
			t.Error("the Bear did not gain swampwalk")
		}
	})
}

func TestOldB789SongOfBloodPumpsAttackersByTheCreatureCardsMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 2; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	}
	for i := 0; i < 2; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	bear := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.P1g2Cast(t, g, "Song of Blood", "Sorcery", oldB789Song, "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 4 {
		t.Errorf("the attacker has power %d, want 4 (two creature cards milled)", p)
	}
}

func TestOldB789CorrosionRustsAndDestroysTheOpponentsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	corrosion := effectstest.PushCatalogPermanent(g, me.ID, "Corrosion", "Enchantment", oldB789Corrode, false)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cheap Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID,
	})
	dear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dear Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("naming the target opponent: %v", err)
		}
	}
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	effectstest.B06AddMana(me, "C")
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the mana value 1 artifact survived its first rust counter")
	}
	if !g.Battlefield.Contains(dear) {
		t.Fatal("the mana value 3 artifact was destroyed by one rust counter")
	}
	if n := effectstest.CounterOn(g, dear, effects.Old789RustCounter); n != 1 {
		t.Errorf("the surviving artifact has %d rust counters, want 1", n)
	}
	// Corrosion leaving takes every rust counter with it.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(corrosion); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterOn(g, dear, effects.Old789RustCounter); n != 0 {
		t.Errorf("rust counters left after Corrosion left: %d", n)
	}
}
