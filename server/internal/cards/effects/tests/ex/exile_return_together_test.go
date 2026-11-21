package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_return_together_test.go — #1872, the exile half of #1867. A
// card that returns several cards from exile to the battlefield puts
// them there in ONE event, so each newcomer sees the others enter
// (CR 603.6a). The probe is graveyard_return_together_test.go's: two
// Soul Wardens returned together gain 2 life, one at a time 1.

// etWarden puts a Soul Warden card owned by p into exile.
func etWarden(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: "Soul Warden", OracleID: effectstest.B04SoulWardenOracle,
		TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// etWardensOnBattlefield counts the Soul Wardens `controller` controls.
func etWardensOnBattlefield(g *game.Game, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Soul Warden" && c.Controller == controller {
				n++
			}
		}
	})
	return n
}

// Living Death's last step is simultaneous for every player: all the
// exiled creature cards enter together.
func TestLivingDeathPutsEveryExiledCreatureOntoTheBattlefieldTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.GtWarden(me)
	effectstest.GtWarden(me)
	effectstest.GtWarden(opp)
	effectstest.CastCatalogSpell(t, g, "Living Death", "Sorcery", effectstest.B03LivingDeathOracle, nil)
	myLife, oppLife := me.Life, opp.Life
	effectstest.GtSettle(t, g)
	if etWardensOnBattlefield(g, me.ID) != 2 || etWardensOnBattlefield(g, opp.ID) != 1 {
		t.Fatal("every Warden returns under its owner's control")
	}
	// My two Wardens each see the other two enter; theirs sees my two.
	if got := me.Life - myLife; got != 4 {
		t.Errorf("I gained %d life, want 4: each of my Wardens sees the other two enter (CR 603.6a)", got)
	}
	if got := opp.Life - oppLife; got != 2 {
		t.Errorf("they gained %d life, want 2: their Warden sees my two enter (CR 603.6a)", got)
	}
}

// The delayed blink's return (Waterbender's Restoration, The Eternal
// Wanderer, Charming Prince, Palace Jailer, Cosmic Intervention) brings
// every card back together.
func TestWaterbendersRestorationReturnsItsCreaturesTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushFlickerCreature(g, me.ID, "Soul Warden", effectstest.B04SoulWardenOracle)
	b := effectstest.PushFlickerCreature(g, me.ID, "Soul Warden", effectstest.B04SoulWardenOracle)
	helpers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		XValue: 2, TapIDs: helpers,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	}); err != nil {
		t.Fatalf("CastSpell Waterbender's Restoration: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, a) || !effectstest.ExileHas(g, b) {
		t.Fatal("both Wardens are exiled")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	life := me.Life
	effectstest.GtSettle(t, g)
	if etWardensOnBattlefield(g, me.ID) != 2 {
		t.Fatal("both Wardens return")
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

// The delayed return skips a card that is no longer in exile
// (CR 603.7c) and a token, which can't come back (CR 111.8), and still
// returns the rest together.
func TestDelayedExileReturnSkipsATokenAndACardThatMovedOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a, b := etWarden(g, me), etWarden(g, me)
	token := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: token, Name: "Soul Warden", OracleID: effectstest.B04SoulWardenOracle,
		TypeLine: "Token Creature — Human Cleric", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	gone := effectstest.GtWarden(me)
	item := &game.StackItem{Controller: me.ID, Targets: effectstest.SgCards(a, token, gone, b)}
	life := me.Life
	g.WithWriteLock(func() {
		if err := effects.ReturnExiledCardsToOwners(g, item); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	effectstest.GtSettle(t, g)
	if etWardensOnBattlefield(g, me.ID) != 2 {
		t.Fatalf("%d Wardens on the battlefield, want the two cards", etWardensOnBattlefield(g, me.ID))
	}
	if effects.OnBattlefield(g, gone) {
		t.Error("a card that left exile is not followed (CR 603.7c)")
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

// Oblivion Sower's lands enter together. The lands are land creatures
// carrying Soul Warden's ability (a Dryad Arbor that is also a Warden),
// so each one that sees another enter gains a life: three together is
// 6 life, one at a time 3.
func TestOblivionSowerPutsTheLandsOntoTheBattlefieldTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := func(name string) game.Card {
		return game.Card{Name: name, OracleID: effectstest.B04SoulWardenOracle,
			TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1}
	}
	lib := effectstest.SeedSearchLibrary(opp, land("Arbor One"), land("Arbor Two"),
		game.Card{Name: "Their Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Their Bolt", TypeLine: "Instant"},
	)
	third := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: third, Name: "Arbor Three", OracleID: effectstest.B04SoulWardenOracle,
		TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	// CR 406.3a: a card exiled face down has no characteristics, so it
	// is not a land card the Sower can take.
	hidden := game.Card{InstanceID: uuid.New(), Name: "Hidden Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID}
	hidden.SetFaceDown(game.FaceDownExiled)
	g.Exile.PushTop(hidden)
	item := &game.StackItem{Controller: me.ID, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	life := me.Life
	g.WithWriteLock(func() {
		if err := effects.B21ExileTopFourThenTakeTheirLands(g, item); err != nil {
			t.Fatalf("Oblivion Sower: %v", err)
		}
	})
	effectstest.EtChooseCards(t, g, me.ID, 3, lib[0], lib[1], third)
	effectstest.GtSettle(t, g)
	for _, name := range []string{"Arbor One", "Arbor Two", "Arbor Three"} {
		id := effectstest.FindBattlefieldByName(g, name)
		if id == uuid.Nil || effectstest.ControllerOf(t, g, id) != me.ID {
			t.Errorf("%s enters under your control", name)
		}
	}
	if got := me.Life - life; got != 6 {
		t.Errorf("gained %d life, want 6: each land sees the other two enter (CR 603.6a)", got)
	}
	if !effectstest.ExileHas(g, hidden.InstanceID) {
		t.Error("a face-down card in exile is not a land card (CR 406.3a)")
	}
}

// An "until this leaves the battlefield" exile that took two cards
// (Hostage Taker's entry trigger doubled by Panharmonicon) gives both
// back together, under their owner's control (CR 610.3c).
func TestUntilLeavesReturnsEveryExiledCardTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taker := effectstest.B12Push(g, me.ID, "Hostage Taker", "Creature — Human Pirate", "", 2, 3)
	a := effectstest.PushFlickerCreature(g, opp.ID, "Soul Warden", effectstest.B04SoulWardenOracle)
	b := effectstest.PushFlickerCreature(g, opp.ID, "Soul Warden", effectstest.B04SoulWardenOracle)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b} {
			g.EmitEvent(game.Event{Kind: game.EventResolve, Source: taker, Label: cardsh.HostageTakerExileLabel})
			if err := g.ExileCardForEffect(id); err != nil {
				t.Fatalf("exile: %v", err)
			}
		}
	})
	if !effectstest.ExileHas(g, a) || !effectstest.ExileHas(g, b) {
		t.Fatal("both Wardens are exiled with the Taker")
	}
	life := opp.Life
	item := &game.StackItem{Controller: me.ID, SourceCardID: taker}
	g.WithWriteLock(func() {
		if err := effects.B41ReturnCardsExiledWithToTheBattlefield(cardsh.HostageTakerExileLabel)(g, item); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	effectstest.GtSettle(t, g)
	if etWardensOnBattlefield(g, opp.ID) != 2 {
		t.Fatal("both Wardens return under their owner's control (CR 610.3c)")
	}
	if got := opp.Life - life; got != 2 {
		t.Errorf("their owner gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

// Teval's land returns tapped as it enters, not untapped and tapped a
// moment later.
func TestTevalReturnsItsLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.Batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	item := &game.StackItem{Controller: me.ID}
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := effects.B25MillThreeThenMayReturnLandTapped(g, item); err != nil {
			t.Fatalf("Teval: %v", err)
		}
	})
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	c, ok := g.LookupCardForEffect(land)
	if !ok || !effects.OnBattlefield(g, land) || !c.Tapped {
		t.Fatal("the land is on the battlefield, tapped")
	}
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[before:] {
			if ev.Kind == game.EventTapCard && ev.CardID == land {
				t.Error("the land was tapped after it entered; it enters tapped")
			}
		}
	})
}
