package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4268 (CR 101.4, CR 616.1): the cards that act on every library, or on every
// chosen player, do it one player at a time in APNAP order. Each test makes two
// libraries ask a CR 616.1 ordering question and checks that exactly one prompt
// is open at a time, the active player's first, and that whatever follows the
// last library waits for it.

// activeAt makes seat i the active player and returns the seats from there.
func activeAt(g *game.Game, i int) {
	g.WithWriteLock(func() { g.Turn.ActiveSeat = i })
}

func hasHandBan(g *game.Game) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if m.Kind == game.ModCantPlayFromHand {
					found = true
				}
			}
		}
	})
	return found
}

func TestEtaliPrimalStormExilesEachLibraryInApnapOrderOnePromptAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 2)
	controller, active, other := g.Seats[1], g.Seats[2], g.Seats[0]
	activeCard := promptingTop(g, active, false)
	otherCard := promptingTop(g, other, false)
	plainBefore := g.Seats[3].Library.Size()

	g.WithWriteLock(func() {
		if err := cardse.EtaliExileAndOfferCasts(g, &game.StackItem{Controller: controller.ID}); err != nil {
			t.Fatalf("etaliExileAndOfferCasts: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the trigger")
	if g.PendingChoices[0].Chooser != active.ID {
		t.Fatal("the first prompt is not the active player's")
	}
	if !stillInLibraryOf(other, otherCard) {
		t.Fatal("a second library was exiled from while the first prompt was open")
	}
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.FindInZone(g.Exile, activeCard); !ok {
		t.Fatal("the active player's card is not exiled after their answer")
	}
	expectOnePrompt(t, g, "after the first answer")
	if g.PendingChoices[0].Chooser != other.ID {
		t.Fatal("the second prompt is not the next player's")
	}
	effectstest.AnswerExileOrder(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts still open", len(g.PendingChoices))
	}
	if got := g.Seats[3].Library.Size(); got != plainBefore-1 {
		t.Errorf("a library that asked nothing lost %d cards, want 1", plainBefore-got)
	}
	// Etali's controller, not the library's owner, may cast each exiled card.
	for _, id := range []uuid.UUID{activeCard, otherCard} {
		if effectstest.PermissionFor(g, controller.ID, id) == nil {
			t.Errorf("Etali's controller has no permission over %s", id)
		}
		if effectstest.PermissionFor(g, g.Seats[3].ID, id) != nil {
			t.Errorf("a bystander has a permission over %s", id)
		}
	}
}

func TestMemoryVesselHandBanWaitsForTheLastLibraryAndEachPlayerHoldsTheirOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 3)
	me, active, other := g.Seats[0], g.Seats[3], g.Seats[1]
	activeCard := promptingTop(g, active, false)
	otherCard := promptingTop(g, other, false)
	bystanderBefore := g.Seats[2].Library.Size()

	spec, _ := effects.Lookup(effectstest.MemoryVesselOracle)
	g.WithWriteLock(func() {
		if err := spec.Activated[0].Effect(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}); err != nil {
			t.Fatalf("Memory Vessel: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the ability")
	if g.PendingChoices[0].Chooser != active.ID {
		t.Fatal("the first prompt is not the active player's")
	}
	effectstest.AnswerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	if g.PendingChoices[0].Chooser != other.ID {
		t.Fatal("the second prompt is not the next player's")
	}
	if hasHandBan(g) {
		t.Fatal("the hand ban was written before the last library was exiled from")
	}
	effectstest.AnswerExileOrder(t, g)
	if !hasHandBan(g) {
		t.Fatal("the hand ban was not written after the last library")
	}
	if effectstest.PermissionFor(g, active.ID, activeCard) == nil || effectstest.PermissionFor(g, other.ID, otherCard) == nil {
		t.Error("a player does not hold the permission over the card they exiled")
	}
	if effectstest.PermissionFor(g, me.ID, activeCard) != nil || effectstest.PermissionFor(g, active.ID, otherCard) != nil {
		t.Error("a player may play a card another player exiled")
	}
	// Seven per library: the six unremarkable cards exile without a question.
	if got := g.Seats[2].Library.Size(); got != bystanderBefore-7 {
		t.Errorf("a library that asked nothing is at %d, want %d", got, bystanderBefore-7)
	}
}

func TestRoccoStreetChefExilesEachLibraryInTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 1)
	active, other := g.Seats[1], g.Seats[2]
	activeCard := promptingTop(g, active, false)
	otherCard := promptingTop(g, other, false)

	spec, _ := effects.Lookup(effectstest.RoccoOracle)
	g.WithWriteLock(func() {
		if err := spec.Triggered[0].Effect(g, &game.StackItem{Controller: g.Seats[0].ID}); err != nil {
			t.Fatalf("Rocco: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the trigger")
	if g.PendingChoices[0].Chooser != active.ID || !stillInLibraryOf(other, otherCard) {
		t.Fatal("the first prompt is not the active player's, or the next library was touched")
	}
	effectstest.AnswerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	effectstest.AnswerExileOrder(t, g)
	if effectstest.PermissionFor(g, active.ID, activeCard) == nil || effectstest.PermissionFor(g, other.ID, otherCard) == nil {
		t.Error("a player does not hold the permission over the card they exiled")
	}
}

func TestDreadSummonsMillsInApnapOrderAndMakesZombiesAfterTheLastMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 0)
	caster, active := g.Seats[1], g.Seats[0]
	casterCard := promptingTop(g, caster, true)
	activeCard := promptingTop(g, active, true)
	zombies := func() int { return len(effectstest.BattlefieldIDsNamed(g, "Zombie")) }

	spec, _ := effects.Lookup(effectstest.B35DreadSummonsOracle)
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: caster.ID, XValue: 1}
		if err := spec.OnResolve(item, effects.NewContext(g, item)); err != nil {
			t.Fatalf("Dread Summons: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the spell")
	if g.PendingChoices[0].Chooser != active.ID {
		t.Fatal("the first prompt is not the active player's: the caster does not go first")
	}
	if !stillInLibraryOf(caster, casterCard) {
		t.Fatal("the caster milled while the active player's prompt was open")
	}
	effectstest.AnswerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	if g.PendingChoices[0].Chooser != caster.ID || zombies() != 0 {
		t.Fatal("the second prompt is not the caster's, or Zombies came before the last mill")
	}
	effectstest.AnswerExileOrder(t, g)
	if !milledInto(active, activeCard) || !milledInto(caster, casterCard) {
		t.Fatal("a prompting card is not milled")
	}
	if got := zombies(); got != 2 {
		t.Fatalf("%d Zombies, want 2 (one per creature card milled)", got)
	}
}

func TestCourtOfCunningMillsEachTargetInApnapOrderOnePromptAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 2)
	owner, active, other := g.Seats[1], g.Seats[2], g.Seats[0]
	activeCard := promptingTop(g, active, true)
	otherCard := promptingTop(g, other, true)

	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: owner.ID, Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: other.ID}, {Kind: game.TargetPlayer, ID: active.ID},
		}}
		if err := cardsco.CourtOfCunningUpkeep(g, item); err != nil {
			t.Fatalf("courtOfCunningUpkeep: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the upkeep trigger")
	if g.PendingChoices[0].Chooser != active.ID {
		t.Fatal("the first prompt is not the active player's, though the other target was named first")
	}
	if !stillInLibraryOf(other, otherCard) {
		t.Fatal("the second target milled while the first prompt was open")
	}
	effectstest.AnswerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	effectstest.AnswerExileOrder(t, g)
	if !milledInto(active, activeCard) || !milledInto(other, otherCard) {
		t.Fatal("a target's prompting card is not milled")
	}
}

func TestConsumingAberrationMillsEachOpponentInApnapOrderOnePromptAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 3)
	controller, active, other := g.Seats[0], g.Seats[3], g.Seats[1]
	activeCard := promptingTop(g, active, true)
	otherCard := promptingTop(g, other, true)

	g.WithWriteLock(func() {
		if err := effects.B14EachOpponentMillsUntilLand(g, &game.StackItem{Controller: controller.ID}); err != nil {
			t.Fatalf("b14EachOpponentMillsUntilLand: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the trigger")
	if g.PendingChoices[0].Chooser != active.ID {
		t.Fatal("the first prompt is not the active player's")
	}
	if !stillInLibraryOf(other, otherCard) {
		t.Fatal("the second opponent milled while the first prompt was open")
	}
	effectstest.AnswerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	effectstest.AnswerExileOrder(t, g)
	if !milledInto(active, activeCard) || !milledInto(other, otherCard) {
		t.Fatal("an opponent's prompting card is not milled")
	}
}

func TestTashasHideousLaughterExilesEachOpponentInApnapOrderOnePromptAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 3)
	controller, active, other := g.Seats[0], g.Seats[3], g.Seats[1]
	activeCard := promptingTop(g, active, false)
	otherCard := promptingTop(g, other, false)

	spec, _ := effects.Lookup("e352f5b9-6406-4914-bc79-f24608be6bc9")
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: controller.ID}
		if err := spec.OnResolve(item, effects.NewContext(g, item)); err != nil {
			t.Fatalf("Tasha's Hideous Laughter: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the spell")
	if g.PendingChoices[0].Chooser != active.ID {
		t.Fatal("the first prompt is not the active player's")
	}
	if !stillInLibraryOf(other, otherCard) {
		t.Fatal("the second opponent was exiled from while the first prompt was open")
	}
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.FindInZone(g.Exile, activeCard); !ok {
		t.Fatal("the active opponent's card is not exiled after their answer")
	}
	expectOnePrompt(t, g, "after the first answer")
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.FindInZone(g.Exile, otherCard); !ok {
		t.Fatal("the second opponent's card is not exiled")
	}
}

// Locke, Treasure Hunter's Mug already mills one library at a time; #4268 puts
// the order right: the active player's library first, not seat zero's.
func TestLockeTreasureHunterMugMillsInApnapOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeAt(g, 2)
	controller, active, other := g.Seats[1], g.Seats[2], g.Seats[0]
	activeCard := promptingTop(g, active, true)
	otherCard := promptingTop(g, other, true)

	g.WithWriteLock(func() {
		if err := cardskr.LockeMug(g, &game.StackItem{Controller: controller.ID}); err != nil {
			t.Fatalf("lockeMug: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the trigger")
	if g.PendingChoices[0].Chooser != active.ID {
		t.Fatal("the first prompt is not the active player's")
	}
	if !stillInLibraryOf(other, otherCard) {
		t.Fatal("seat zero milled before the active player")
	}
	effectstest.AnswerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	effectstest.AnswerExileOrder(t, g)
	if !milledInto(active, activeCard) || !milledInto(other, otherCard) {
		t.Fatal("a prompting card is not milled")
	}
}
