package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// s82_old_b266_cards_test.go — S82 OLD batch 266 (Kaladesh): Trusty Companion,
// Inventor's Goggles, Sequestered Stash. Each test plays the catalog card
// through the engine.

const (
	oldB266Companion = "bd8c633c-ba99-47ac-a1d1-9e15aa948933"
	oldB266Goggles   = "ee32ab1a-1120-4c34-a538-7bde9eff5a6e"
	oldB266Stash     = "b6fe779f-b20d-49cc-96dd-54f1ffb312e1"
)

func TestOldB266EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB266Companion: "Trusty Companion", oldB266Goggles: "Inventor's Goggles", oldB266Stash: "Sequestered Stash",
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

func TestOldB266TrustyCompanionCantAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dog := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Trusty Companion", oldB266Companion, 3, 2))
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)

	if err := g.DeclareAttacker(dog, opp.ID); err != nil {
		t.Fatalf("declaring the Companion first was refused: %v", err)
	}
	pass := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindPass {
			pass = true
		}
	}
	if pass {
		t.Error("the pass is offered while the Companion attacks alone")
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("a second attacker was refused: %v", err)
	}
	if got := effectstest.EffectiveAbilities(t, g, dog); !effectstest.ContainsString(got, "vigilance") {
		t.Errorf("abilities %v lack vigilance", got)
	}
}

func TestOldB266InventorsGogglesPumpAndAttachToAnArtificer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	goggles := effectstest.SeedEquipment(g, me.ID, "Inventor's Goggles", oldB266Goggles)
	effectstest.EquipTo(t, g, me.ID, goggles, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 4 {
		t.Fatalf("equipped Bear is %d/%d, want 3/4", p, tough)
	}

	smithCard := effectstest.ApaCreature("Smith", "", 1, 1)
	smithCard.TypeLine = "Creature — Human Artificer"
	smith := effectstest.ApaPush(g, me.ID, me.ID, smithCard)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: smith})
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, goggles); host.ID != smith {
		t.Fatalf("Goggles host = %+v, want the Artificer", host)
	}
}

func TestOldB266InventorsGogglesIgnoreANonArtificerAndOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	goggles := effectstest.SeedEquipment(g, me.ID, "Inventor's Goggles", oldB266Goggles)
	plain := effectstest.SeedBear(g, me.ID)
	theirCard := effectstest.ApaCreature("Theirs", "", 1, 1)
	theirCard.TypeLine = "Creature — Human Artificer"
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, theirCard)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: plain})
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: theirs})
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("the Goggles offered to attach to a non-Artificer or an opponent's Artificer")
		}
	}
	if host := effectstest.AttachmentHostOf(t, g, goggles); host.ID != uuid.Nil {
		t.Errorf("Goggles attached to %+v", host)
	}
}

func oldB266Stash5(t *testing.T) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID, uuid.UUID) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	stash := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Sequestered Stash", TypeLine: "Land", OracleID: oldB266Stash})
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Milled Rock", TypeLine: "Artifact"},
		game.Card{Name: "Milled Bolt", TypeLine: "Instant"},
		game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Milled Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Milled Wrath", TypeLine: "Sorcery"},
		game.Card{Name: "Stays", TypeLine: "Creature — Bear"})
	return g, me, stash, ids, old
}

func TestOldB266SequesteredStashMillsFiveThenPutsAnyGraveyardArtifactOnTop(t *testing.T) {
	g, me, stash, ids, old := oldB266Stash5(t)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.ApaActivateNoPoolCheck(t, g, me, stash, 0)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for an artifact card")
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("offered %v, want the old and the milled artifact", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, old)
	if top, err := me.Library.Top(); err != nil || top.InstanceID != old {
		t.Error("the old artifact is not on top of the library")
	}
	if effectstest.FindBattlefieldCardForTest(g, stash) != nil {
		t.Error("the land was not sacrificed")
	}
	for _, i := range []int{0, 1, 2, 3, 4} {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("milled card %d is not in the graveyard", i)
		}
	}
}

func TestOldB266SequesteredStashMayDeclineAndNeedsFourMana(t *testing.T) {
	g, me, stash, _, _ := oldB266Stash5(t)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, stash, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with three mana")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.ApaActivateNoPoolCheck(t, g, me, stash, 0)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if me.Library.Size() != 1 {
		t.Errorf("library holds %d, want only the unmilled card", me.Library.Size())
	}
}
