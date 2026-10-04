package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undying_persist_pool3_test.go — the third pool PR after #2075 (ADR
// 0113 §4, owner decision 2). One test per card.

const (
	p3FurystokeGiant  = "0cc62807-aeb4-4dfb-b001-4e40bf703497"
	p3PuppeteerClique = "00e0b103-892c-49b3-836d-867fff197bbd"
	p3WitchKing       = "84137013-7a23-4bc8-b855-04889373fdae"
	p3HauntedOne      = "689cd6a8-8be0-49a6-9758-a81cc9f55cc8"
)

func TestUndyingPersistPool3Registered(t *testing.T) {
	for _, oracle := range []string{p3FurystokeGiant, p3PuppeteerClique, p3WitchKing, p3HauntedOne} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered %v, completeness %v", oracle, ok, spec.Completeness)
		}
	}
}

func p3HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	has := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}

// Furystoke Giant: another creature gains the {T}: 2 damage ability.
func TestFurystokeGiantGrantsThePing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	enterCard(t, g, me.ID, game.Card{Name: "Furystoke Giant", TypeLine: "Creature — Giant Warrior", OracleID: p3FurystokeGiant, Power: 3, Toughness: 3})
	settleProwess(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("the bear's granted ability: %v", err)
	}
	settleProwess(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2", life, opp.Life)
	}
}

// Puppeteer Clique: an opponent's creature card comes back under my
// control with haste and is exiled at my next end step.
func TestPuppeteerCliqueBorrowsAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	enterCard(t, g, me.ID, game.Card{Name: "Puppeteer Clique", TypeLine: "Creature — Faerie Wizard", OracleID: p3PuppeteerClique, Power: 3, Toughness: 2})
	b04WaitForPick(t, g, me.ID)
	for _, id := range latestPickTarget(g, me.ID).PickTargetCards {
		if id == mine {
			t.Error("a card from my own graveyard was offered")
		}
	}
	pickCard(t, g, me.ID, theirs)
	settleProwess(t, g)
	c, ok := battlefieldCard(g, theirs)
	if !ok || c.Controller != me.ID {
		t.Fatal("the creature is not on the battlefield under my control")
	}
	if !p3HasKeyword(g, theirs, "haste") {
		t.Error("it has no haste")
	}
	advanceTo(t, g, game.StepEnd)
	settleProwess(t, g)
	if onBattlefield(g, theirs) || !g.Exile.Contains(theirs) {
		t.Error("it was not exiled at my end step")
	}
}

// Witch-king: attacking with a Wraith exiles its power in cards, playable.
func TestWitchKingImpulsesTheWraithsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wk := poolPush(g, me.ID, "Witch-king, Sky Scourge", p3WitchKing, 5, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wk {
				g.Battlefield.Cards[i].TypeLine = "Legendary Creature — Wraith Noble"
			}
		}
	})
	lib, ex := me.Library.Size(), g.Exile.Size()
	attackWith(t, g, opp.ID, wk)
	settleProwess(t, g)
	if got := lib - me.Library.Size(); got < 5 {
		t.Errorf("%d cards left the library, want 5 (plus any draws)", got)
	}
	if got := g.Exile.Size() - ex; got != 5 {
		t.Errorf("%d cards exiled, want 5", got)
	}
}

// Haunted One: my commander becoming tapped gives it and creatures that
// share a type with it +2/+0 and undying.
func TestHauntedOneGivesTheTribeUndying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Haunted One", TypeLine: "Legendary Enchantment — Background",
		OracleID: p3HauntedOne, Owner: me.ID, Controller: me.ID})
	cmdr := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Zombie Lord", TypeLine: "Legendary Creature — Zombie",
		Power: 2, Toughness: 2, IsCommander: true, Owner: me.ID, Controller: me.ID})
	zombie := b16Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, cmdr)
	settleProwess(t, g)
	if !p3HasKeyword(g, zombie, game.KeywordUndying) || !p3HasKeyword(g, cmdr, game.KeywordUndying) {
		t.Error("the commander and the Zombie should have undying")
	}
	if p3HasKeyword(g, bear, game.KeywordUndying) {
		t.Error("the Bear shares no type and should not")
	}
	if p := powerOf(t, g, zombie); p != 4 {
		t.Errorf("Zombie power %d, want 4", p)
	}
}
