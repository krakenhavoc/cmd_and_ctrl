package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b305_cards_test.go — Battle for Zendikar Devoid cards, batch 305
// (#5547). Each test plays the real catalog entry and asserts what its text
// says.

const (
	oldB305MindRaker   = "060076e7-3705-4f8e-8fac-59402835902b"
	oldB305Feeder      = "0767fb65-a621-494b-97a4-f23a4c8b4bec"
	oldB305Incubator   = "088d63ec-bd5e-4e80-acf9-2af270a9b395"
	oldB305Reclaimer   = "137de343-e3d0-4ccb-b8fc-680024387367"
	oldB305Strangler   = "197f3b17-ce21-430e-8e41-f5d02d5e3637"
	oldB305Forerunner  = "1c3b5d50-6b5c-49a3-b64b-5bee51d22725"
	oldB305FromBeyond  = "31b27775-196b-4022-bff5-25b34290ff7f"
	oldB305Guide       = "3394ce74-c352-4ae1-a7bb-f6ef8ac1c810"
	oldB305Strider     = "386b8c1f-620c-4561-a9d5-d01e2eb5f12a"
	oldB305Dominator   = "3c98c925-f380-4520-970d-4c1abbebefe3"
	oldB305TideDrifter = "428c18aa-b195-4cf2-a1a9-09723b26eb0a"
)

func TestOldB305CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB305MindRaker: "Mind Raker", oldB305Feeder: "Fathom Feeder", oldB305Incubator: "Incubator Drone",
		oldB305Reclaimer: "Ulamog's Reclaimer", oldB305Strangler: "Wasteland Strangler",
		oldB305Forerunner: "Forerunner of Slaughter", oldB305FromBeyond: "From Beyond",
		oldB305Guide: "Ruination Guide", oldB305Strider: "Murk Strider", oldB305Dominator: "Dominator Drone",
		oldB305TideDrifter: "Tide Drifter",
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

// oldB305Exile puts a card `owner` owns into exile.
func oldB305Exile(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled Card", TypeLine: "Creature — Bear", Owner: owner, Controller: owner})
	return id
}

// oldB305Colored puts a coloured creature onto the battlefield.
func oldB305Colored(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: "{G}", Colors: []string{"G"},
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// oldB305CastProcessor casts the creature, picks the trigger's target when it
// has one, and answers the "you may" yes; it returns whether the pick prompt
// for the exiled card appeared, answering it with `pick`.
func oldB305CastProcessor(t *testing.T, g *game.Game, me *game.Player, name, oracle string, target uuid.UUID, answer bool, pick uuid.UUID) {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, name, "Creature — Eldrazi Processor", oracle)
	effectstest.PassPriorityAroundTable(t, g)
	if target != uuid.Nil {
		effectstest.PickTriggerTarget(t, g, me.ID, target)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, answer)
	if !answer {
		effectstest.PassPriorityAroundTable(t, g)
		return
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt for the exiled card")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB305MindRakerMakesEachOpponentDiscardOnlyIfACardWasProcessed(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		exiled := oldB305Exile(g, opp.ID)
		keep := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: keep, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		hand := opp.Hand.Size()
		oldB305CastProcessor(t, g, me, "Mind Raker", oldB305MindRaker, uuid.Nil, yes, exiled)
		if yes {
			if opp.Graveyard.Size() < 1 || !opp.Graveyard.Contains(exiled) {
				t.Error("the exiled card did not go to its owner's graveyard")
			}
			if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
				effectstest.AnswerDiscard(t, g, opp.ID, keep)
				effectstest.PassPriorityAroundTable(t, g)
			}
			if opp.Hand.Size() != hand-1 {
				t.Errorf("opponent hand %d, want %d", opp.Hand.Size(), hand-1)
			}
		} else if opp.Hand.Size() != hand || opp.Graveyard.Contains(exiled) {
			t.Errorf("declined: hand %d (want %d), card processed %v", opp.Hand.Size(), hand, opp.Graveyard.Contains(exiled))
		}
	}
}

func TestOldB305ProcessorAsksNothingWithoutAnOpponentsExiledCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB305Exile(g, me.ID) // my own card does not count
	effectstest.CastAndResolveCreature(t, g, "Mind Raker", "Creature — Eldrazi Processor", oldB305MindRaker)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked to process with no card an opponent owns in exile")
	}
}

func TestOldB305UlamogsReclaimerReturnsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exiled := oldB305Exile(g, opp.ID)
	bolt := effectstest.PushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
	effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	oldB305CastProcessor(t, g, me, "Ulamog's Reclaimer", oldB305Reclaimer, bolt, true, exiled)
	if !me.Hand.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Error("the instant did not return to hand")
	}
	if !opp.Graveyard.Contains(exiled) {
		t.Error("the exiled card did not go to its owner's graveyard")
	}
}

func TestOldB305UlamogsReclaimerDeclinedReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exiled := oldB305Exile(g, opp.ID)
	bolt := effectstest.PushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
	oldB305CastProcessor(t, g, me, "Ulamog's Reclaimer", oldB305Reclaimer, bolt, false, exiled)
	if me.Hand.Contains(bolt) {
		t.Error("returned the instant without processing a card")
	}
}

func TestOldB305WastelandStranglerShrinksATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exiled := oldB305Exile(g, opp.ID)
	victim := oldB305Colored(g, opp.ID, "Their Bear", 4, 4)
	oldB305CastProcessor(t, g, me, "Wasteland Strangler", oldB305Strangler, victim, true, exiled)
	if p, tough := effectstest.EffectivePower(t, g, victim), effectstest.EffectiveToughness(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("target is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB305WastelandStranglerDeclinedLeavesTheTargetAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exiled := oldB305Exile(g, opp.ID)
	victim := oldB305Colored(g, opp.ID, "Their Bear", 4, 4)
	oldB305CastProcessor(t, g, me, "Wasteland Strangler", oldB305Strangler, victim, false, exiled)
	if effectstest.EffectivePower(t, g, victim) != 4 {
		t.Error("the creature shrank without a processed card")
	}
}

func TestOldB305MurkStriderBouncesATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exiled := oldB305Exile(g, opp.ID)
	victim := oldB305Colored(g, opp.ID, "Their Bear", 2, 2)
	oldB305CastProcessor(t, g, me, "Murk Strider", oldB305Strider, victim, true, exiled)
	if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
		t.Error("the creature was not returned to its owner's hand")
	}
}

func TestOldB305IncubatorDroneMakesAScion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := countScions(g, me.ID)
	effectstest.CastAndResolveCreature(t, g, "Incubator Drone", "Creature — Eldrazi Drone", oldB305Incubator)
	effectstest.PassPriorityAroundTable(t, g)
	if got := countScions(g, me.ID); got != before+1 {
		t.Errorf("%d Scions after Incubator Drone, want %d", got, before+1)
	}
}

func countScions(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsToken() && c.HasSubtype("Eldrazi") && c.HasSubtype("Scion") {
			n++
		}
	}
	return n
}

func TestOldB305FromBeyondMakesAScionEachUpkeepAndFetchesAnEldrazi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fb := effectstest.B12Push(g, me.ID, "From Beyond", "Enchantment", oldB305FromBeyond, 0, 0)
	before := countScions(g, me.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := countScions(g, me.ID); got != before+1 {
		t.Errorf("%d Scions after the upkeep, want %d", got, before+1)
	}
	// the sacrifice-and-search ability
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Big Eldrazi", TypeLine: "Creature — Eldrazi", Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CG")
	if err := g.ActivateCatalogAbility(me.ID, fb, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fb) {
		t.Error("From Beyond was not sacrificed")
	}
	if open := effectstest.ChkB1OpenSearch(g); open != nil {
		effectstest.BeginSearchIfAsked(t, g, open)
	}
	effectstest.TakeSearchPrompts(t, g)
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Big Eldrazi" {
			found = true
		}
	}
	if !found {
		t.Error("the Eldrazi card was not put into hand")
	}
}

func TestOldB305ForerunnerGivesAColorlessCreatureHasteOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fr := effectstest.B12Push(g, me.ID, "Forerunner of Slaughter", "Creature — Eldrazi Drone", oldB305Forerunner, 3, 2)
	colored := oldB305Colored(g, me.ID, "Green Bear", 2, 2)
	colorless := effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, fr, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(colored), Strict: true}); err == nil {
		t.Fatal("a coloured creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, fr, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(colorless), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, colorless, "haste") {
		t.Error("the colorless creature did not gain haste")
	}
}

func TestOldB305RuinationGuideAndTideDrifterPumpOtherColorlessCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 2, 2)
	green := oldB305Colored(g, me.ID, "Green Bear", 2, 2)
	guide := effectstest.B12Push(g, me.ID, "Ruination Guide", "Creature — Eldrazi Drone", oldB305Guide, 3, 2)
	drifter := effectstest.B12Push(g, me.ID, "Tide Drifter", "Creature — Eldrazi Drone", oldB305TideDrifter, 0, 5)
	if p, tough := effectstest.EffectivePower(t, g, golem), effectstest.EffectiveToughness(t, g, golem); p != 3 || tough != 3 {
		t.Errorf("colorless Golem is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, green), effectstest.EffectiveToughness(t, g, green); p != 2 || tough != 2 {
		t.Errorf("green Bear is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, guide), effectstest.EffectiveToughness(t, g, guide); p != 3 || tough != 3 {
		t.Errorf("Guide is %d/%d, want 3/3 (Tide Drifter's +0/+1, not its own +1/+0)", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, drifter), effectstest.EffectiveToughness(t, g, drifter); p != 1 || tough != 5 {
		t.Errorf("Tide Drifter is %d/%d, want 1/5 (the Guide's +1/+0, not its own +0/+1)", p, tough)
	}
}

func TestOldB305IngestExilesTheTopCardOfTheDamagedPlayersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	guide := effectstest.B12Push(g, me.ID, "Ruination Guide", "Creature — Eldrazi Drone", oldB305Guide, 3, 2)
	ids := effectstest.SeedLibrary(opp, "a", "b", "c")
	lib := opp.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, guide)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card exiled", lib, opp.Library.Size())
	}
	if !g.Exile.Contains(ids[len(ids)-1]) && !g.Exile.Contains(ids[0]) {
		t.Error("the top card is not in exile")
	}
}

func TestOldB305FathomFeederDrawsAndExilesFromEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	feeder := effectstest.B12Push(g, me.ID, "Fathom Feeder", "Creature — Eldrazi Drone", oldB305Feeder, 1, 1)
	effectstest.SeedLibrary(me, "x", "y")
	for _, s := range g.Seats[1:] {
		effectstest.SeedLibrary(s, "a", "b", "c")
	}
	hand := me.Hand.Size()
	libs := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		libs[s.ID] = s.Library.Size()
	}
	effectstest.FloatForTest(g, me, "CCCUB")
	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn", hand, me.Hand.Size())
	}
	for _, s := range g.Seats[1:] {
		if s.Library.Size() != libs[s.ID]-1 {
			t.Errorf("opponent library %d -> %d, want one exiled", libs[s.ID], s.Library.Size())
		}
	}
}

func TestOldB305DominatorDroneDrainsOnlyWithAnotherColorlessCreature(t *testing.T) {
	for _, other := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if other {
			effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 2, 2)
		} else {
			oldB305Colored(g, me.ID, "Green Bear", 2, 2)
		}
		life := opp.Life
		effectstest.CastAndResolveCreature(t, g, "Dominator Drone", "Creature — Eldrazi Drone", oldB305Dominator)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if other {
			want = life - 2
		}
		if opp.Life != want {
			t.Errorf("another colorless creature=%v: opponent life %d, want %d", other, opp.Life, want)
		}
	}
}
