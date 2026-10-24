package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// one_off_resolution_cast_bot_test.go — #4250, part 2, through the real
// engine, the real enumerator and the real heuristic: the one-off "you may
// cast it" offers ask the resolution_cast prompt inside their resolution,
// from a graveyard (Memory Plunder), from exile with a follow-up after the
// casts (Kefka, Dancing Mad) and from the top of a library (Aetherworks
// Marvel). The seat must answer each (#544), and the engine must accept
// what it chose (botAnswersTheDoor, resolution_door_cast_bot_test.go). Not
// behind AISEAT_GAME_TESTS: a few decisions each.

const (
	botMemoryPlunderOracle     = "36b67014-0507-4976-ad52-a2a36676197d"
	botKefkaOracle             = "fe5690d3-547b-4ce9-8e94-77fdc0e9c5c6"
	botAetherworksMarvelOracle = "470ec98e-120c-446d-ade6-7d80aa836b55"
)

func botPassUntilOwed(t *testing.T, g *game.Game, me uuid.UUID) {
	t.Helper()
	for i := 0; i < 16 && !owesChoice(g, me); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me) {
		t.Fatal("the offer was never asked")
	}
}

// Memory Plunder: a free instant out of an opponent's graveyard.
func TestHeuristicSeatAnswersAGraveyardResolutionCast(t *testing.T) {
	g := newSettledTable(t, 42503)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil
	stolen := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: stolen, Name: "Stolen Instant", TypeLine: "Instant", ManaCost: "{1}{U}",
		Owner: opp.ID, Controller: opp.ID})
	plunder := uuid.New()
	c := game.Card{InstanceID: plunder, Name: "Memory Plunder", OracleID: botMemoryPlunderOracle, TypeLine: "Instant",
		ManaCost: "{U/B}{U/B}{U/B}{U/B}", Owner: me.ID, Controller: me.ID}
	c.AddKnowersAll(seatIDs(g))
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, plunder, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: stolen}},
	}); err != nil {
		t.Fatalf("cast Memory Plunder: %v", err)
	}
	botPassUntilOwed(t, g, me.ID)
	card := botAnswersTheDoor(t, g, me.ID, game.MayCastKeywordFreeCast)
	if !g.Stack.Contains(card) && !opp.Graveyard.Contains(card) {
		t.Error("the offered card was neither cast nor left in the graveyard")
	}
}

// Kefka: the end-step trigger exiles a card from the opponent's graveyard;
// whatever the bot answers, the life loss follows the casts.
func TestHeuristicSeatAnswersKefkasResolutionCast(t *testing.T) {
	g := newSettledTable(t, 42504)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil
	opp.Graveyard.Cards = nil
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Sorcery", TypeLine: "Sorcery", ManaCost: "{3}",
		Owner: opp.ID, Controller: opp.ID})
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Kefka, Dancing Mad", OracleID: botKefkaOracle,
		TypeLine: "Legendary Creature — Human Wizard", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	life := opp.Life
	advanceToStep(t, g, game.StepEnd)
	botPassUntilOwed(t, g, me.ID)
	card := botAnswersTheDoor(t, g, me.ID, game.MayCastKeywordFreeCast)
	switch {
	case g.Stack.Contains(card):
		if opp.Life != life-3 {
			t.Errorf("life %d after the cast, want %d", opp.Life, life-3)
		}
	case g.Exile.Contains(card):
		if opp.Life != life {
			t.Errorf("life %d after the decline, want %d", opp.Life, life)
		}
	default:
		t.Error("the offered card was neither cast nor left in exile")
	}
}

// Aetherworks Marvel: one free spell from among the top six, cast where it
// sits in the library.
func TestHeuristicSeatAnswersALibraryResolutionCast(t *testing.T) {
	g := newSettledTable(t, 42505)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Command.Cards = nil
	me.Library.Cards = nil
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	}
	wurm := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: wurm, Name: "Huge Wurm", TypeLine: "Creature — Wurm", ManaCost: "{7}{G}",
		Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	marvel := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: marvel, Name: "Aetherworks Marvel", OracleID: botAetherworksMarvelOracle,
		TypeLine: "Legendary Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 6); err != nil {
			t.Fatalf("energy: %v", err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, marvel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	botPassUntilOwed(t, g, me.ID)
	botAnswersTheDoor(t, g, me.ID, game.MayCastKeywordFreeCast)
	if !g.Stack.Contains(wurm) {
		t.Error("the bot did not cast the free eight-drop from the library")
	}
}
