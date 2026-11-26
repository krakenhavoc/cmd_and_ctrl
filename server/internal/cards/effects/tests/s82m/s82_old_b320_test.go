package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b320_test.go — S82 older sets, batch 320: the Magic Origins card
// draw and library cards.

const (
	oldB320JayemdaeTome   = "39ee576a-0803-4063-9c84-f2b537e4d44c"
	oldB320Bellower       = "5859fd59-8732-414a-8932-007980af495d"
	oldB320Dragon         = "9be9d534-470d-463a-ba2f-9817a6723692"
	oldB320GreatAurora    = "b9d7341e-10e7-43f9-bee2-e8053b287acd"
	oldB320Hartebeest     = "c3a1bb65-556a-4016-a06a-3894710cb190"
	oldB320HallowedMoonlt = "e9541db2-2874-4307-b896-c403f6d82e68"
)

func TestOldB320EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB320JayemdaeTome: "Jayemdae Tome", oldB320Bellower: "Woodland Bellower",
		oldB320Dragon: "Avaricious Dragon", oldB320GreatAurora: "The Great Aurora",
		oldB320Hartebeest: "Totem-Guide Hartebeest", oldB320HallowedMoonlt: "Hallowed Moonlight",
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

func TestOldB320JayemdaeTomeDrawsForFourAndTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	book := effectstest.B12Push(g, me.ID, "Jayemdae Tome", "Artifact — Book", oldB320JayemdaeTome, 0, 0)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("three mana paid for {4}")
	}
	effectstest.FloatForTest(g, me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the tapped Tome activated again")
	}
}

func TestOldB320WoodlandBellowerFindsANonlegendaryGreenCreatureOfManaValueThreeOrLess(t *testing.T) {
	green := func(name, typeLine, cost string) game.Card {
		return game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, Colors: []string{"G"}}
	}
	setup := func(t *testing.T) (*game.Game, *game.Player) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedSearchLibrary(me,
			green("Green Three", "Creature — Elf", "{2}{G}"),
			green("Green Four", "Creature — Elf", "{3}{G}"),
			green("Green Legend", "Legendary Creature — Elf", "{1}{G}"),
			game.Card{Name: "Red Two", TypeLine: "Creature — Goblin", ManaCost: "{1}{R}", Colors: []string{"R"}},
			green("Green Spell", "Instant", "{G}"),
		)
		effectstest.CastCatalogSpell(t, g, "Woodland Bellower", "Creature — Beast", oldB320Bellower, nil)
		effectstest.PassPriorityAroundTable(t, g)
		return g, me
	}
	t.Run("only the printed filter is offered", func(t *testing.T) {
		g, me := setup(t)
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("you may search: the chooser asks")
		}
		if effectstest.SearchOptionNamed(g, c, "Green Three") == uuid.Nil {
			t.Error("a nonlegendary green creature with mana value 3 is not offered")
		}
		for _, name := range []string{"Green Four", "Green Legend", "Red Two", "Green Spell"} {
			if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
				t.Errorf("%s is offered", name)
			}
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, "Green Three")
		if _, ok := effectstest.SearchFetchedCard(g, "Green Three"); !ok {
			t.Error("the pick did not enter the battlefield")
		}
	})
	t.Run("finding nothing is allowed", func(t *testing.T) {
		g, me := setup(t)
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
		if _, ok := effectstest.SearchFetchedCard(g, "Green Three"); ok {
			t.Error("a card entered although none was chosen")
		}
	})
}

func TestOldB320HartebeestMayTutorAnAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Rancor", TypeLine: "Enchantment — Aura", ManaCost: "{G}"},
		game.Card{Name: "Glorious Anthem", TypeLine: "Enchantment", ManaCost: "{1}{W}{W}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	effectstest.CastCatalogSpell(t, g, "Totem-Guide Hartebeest", "Creature — Antelope", oldB320Hartebeest, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("you may search: the chooser asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Rancor") == uuid.Nil {
		t.Error("the Aura is not offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Glorious Anthem") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Aura enchantment or a creature is offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Rancor")
	if !effectstest.B02bHandHasNamed(me, "Rancor") {
		t.Error("the pick does not go to hand")
	}
}

func TestOldB320AvariciousDragonDrawsExtraInYourDrawStepAndDiscardsInYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Avaricious Dragon", "Creature — Dragon", oldB320Dragon, 4, 4)
	// An opponent's draw step is not ours.
	oppHand := g.Seats[1].Hand.Size()
	effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NmAdvanceTo(t, g, game.StepPrecombatMain)
	if got := g.Seats[1].Hand.Size(); got != oppHand+1 {
		t.Errorf("the opponent holds %d cards after their draw step, want %d", got, oppHand+1)
	}
	before := me.Hand.Size()
	effectstest.AdvanceToDrawStepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NmAdvanceTo(t, g, game.StepPrecombatMain)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand after our draw step %d, want %d (the draw and one more)", got, before+2)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 0 {
		t.Errorf("hand after our end step %d, want 0", got)
	}
}

func TestOldB320HallowedMoonlightExilesCreaturesThatWereNotCastUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Hallowed Moonlight", "Instant", oldB320HallowedMoonlt, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: one card is drawn", got, hand+1)
	}
	// A creature put onto the battlefield is exiled instead.
	put := effectstest.B13HandCard(opp, "Put Bear", "Creature — Bear", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, put); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if g.Battlefield.Contains(put) || !g.Exile.Contains(put) {
		t.Error("a creature that was put onto the battlefield was not exiled")
	}
	// A creature token is exiled instead too: it never enters.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 red Goblin"), 1)
	})
	if got := len(effectstest.BattlefieldIDsNamed(g, "Goblin")); got != 0 {
		t.Errorf("a creature token entered (%d Goblins)", got)
	}
	// A land and a creature spell that resolves enter as usual.
	land := effectstest.B13HandCard(me, "Forest", "Basic Land — Forest", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, land); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was exiled")
	}
	bear := effectstest.CastCatalogSpell(t, g, "Cast Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature spell that was cast and resolved did not enter")
	}
	// It ends with the turn.
	effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
	later := effectstest.B13HandCard(opp, "Later Bear", "Creature — Bear", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, later); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if !g.Battlefield.Contains(later) {
		t.Error("the replacement outlived the turn")
	}
}

func TestOldB320TheGreatAuroraShufflesHandsAndPermanentsDrawsThatManyAndPutsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
	})
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	effectstest.SeedSearchLibrary(opp)
	effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 red Goblin"), 1)
	})
	effectstest.B13HandCard(me, "Spell A", "Sorcery", "")
	effectstest.B13HandCard(me, "Spell B", "Instant", "")
	effectstest.B13HandCard(me, "Forest", "Basic Land — Forest", "")
	effectstest.B13HandCard(opp, "Their Forest", "Basic Land — Forest", "")
	effectstest.B13HandCard(opp, "Their Spell", "Sorcery", "")
	aurora := effectstest.CastCatalogSpell(t, g, "The Great Aurora", "Sorcery", oldB320GreatAurora, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "My Bear")) != 0 {
		t.Error("a permanent stayed on the battlefield")
	}
	if len(effectstest.BattlefieldIDsNamed(g, "Goblin")) != 0 {
		t.Error("the token stayed on the battlefield")
	}
	if !g.Exile.Contains(aurora) {
		t.Error("The Great Aurora was not exiled")
	}
	// The count was three cards in hand + the Bear + the token = five; the token
	// ceased to exist, so the four real cards and the filler fill the hand.
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no land prompt for the active player")
	}
	if got := me.Hand.Size(); got != 5 {
		t.Errorf("hand %d, want 5 (hand 3 + a Bear + a token)", got)
	}
	if len(c.ChooseCards) != 1 {
		t.Fatalf("only the one land card is offered, got %d options", len(c.ChooseCards))
	}
	forest := c.ChooseCards[0]
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	// The opponent declines.
	if oc := effectstest.ChooseCardsChoiceFor(g, opp.ID); oc == nil {
		t.Fatal("no land prompt for the opponent")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("the picked land did not enter")
	}
	if got := opp.Hand.Size(); got != 2 {
		t.Errorf("opponent hand %d, want 2", got)
	}
	if g.Seats[2].Hand.Size() != 0 {
		t.Errorf("a player with nothing to shuffle drew %d", g.Seats[2].Hand.Size())
	}
}
