package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_a2_test.go — the rest of the first slice of the Terra deck
// request (#3655).

const (
	deck3655AngelOracle     = "14878a5f-7983-4133-a1a7-9d539531e092"
	deck3655SnortOracle     = "ab94985c-5a3e-42c7-be10-29c5b638346b"
	deck3655GrinderOracle   = "8d0445ef-370b-4f4f-8975-a40cb02c031e"
	deck3655SiegfriedOracle = "a7cac6c3-cbb1-41f4-905d-2eb581e8e2e0"
	deck3655LeoOracle       = "088166b8-5b0b-4ba9-9b0d-358ae523f336"
	deck3655LegionsOracle   = "dc29cb25-57f5-4cac-9764-efa644b41296"
	deck3655PhoenixOracle   = "8cc7e418-cc5e-4aef-9757-874223fccd74"
	deck3655KeyOracle       = "a242c9bf-dab6-4996-bb3a-b4356f9ed0c9"
)

func TestDeck3655SiegfriedCountsTheMilledCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sieg := pushCatalogPermanent(g, me.ID, "Siegfried, Famed Swordsman", "Legendary Creature — Human Warrior Rogue", deck3655SiegfriedOracle, false)
	deck3655GraveyardCreature(me, "Dead One", 1)
	deck3655GraveyardCreature(me, "Dead Two", 1)

	spec, _ := Lookup(deck3655SiegfriedOracle)
	item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: sieg}
	if err := spec.Triggered[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	// The three milled filler cards are not creatures: 2 creature cards, 4 counters.
	if got := battlefieldCardCopy(t, g, sieg).Counters[game.CounterPlusOne]; got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", got)
	}
}

func TestDeck3655LeoReturnsASmallCreatureAndCountsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leo := pushCatalogPermanent(g, me.ID, "General Leo Cristophe", "Legendary Creature — Human Soldier", deck3655LeoOracle, false)
	pushCostedPermanentForTest(g, me.ID, "Bystander", "Creature — Bear", "{1}")
	back := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: back, Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{2}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})

	spec, _ := Lookup(deck3655LeoOracle)
	item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: leo,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: back}}}
	if err := spec.Triggered[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	if !g.Battlefield.Contains(back) {
		t.Fatal("the creature card did not return")
	}
	// Leo, the bystander and the returned creature.
	if got := battlefieldCardCopy(t, g, leo).Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

func TestDeck3655LegionsExilesTheTargetAndItsNamesakeTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var bears []uuid.UUID
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 green Bear"), 3)
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 1)
		for _, c := range g.Battlefield.Cards {
			if c.Controller == opp.ID && c.IsToken() {
				bears = append(bears, c.InstanceID)
			}
		}
	})
	if len(bears) != 3 {
		t.Fatalf("setup made %d opposing tokens, want 3", len(bears))
	}
	other := pushCostedPermanentForTest(g, opp.ID, "Other Creature", "Creature — Cat", "{2}")

	castCatalogSpell(t, g, "Legions to Ashes", "Sorcery", deck3655LegionsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bears[0]}})
	passPriorityAroundTable(t, g)

	for _, id := range bears {
		if g.Battlefield.Contains(id) {
			t.Errorf("opposing token %s is still on the battlefield", id)
		}
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a differently named permanent was exiled")
	}
	mine := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() {
			mine++
		}
	}
	if mine != 1 {
		t.Errorf("the caster's own namesake token count = %d, want 1", mine)
	}
}

func TestDeck3655SnortDamagesOnlyOpponentsWhoDiscarded(t *testing.T) {
	g := newCatalogGame(t)
	me, yes, no := g.Seats[0], g.Seats[1], g.Seats[2]
	yesLife, noLife := yes.Life, no.Life
	castCatalogSpell(t, g, "Snort", "Sorcery", deck3655SnortOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		answerMayChoice(t, g, p.ID, p.ID == me.ID || p.ID == yes.ID)
	}
	if got := yesLife - yes.Life; got != 5 {
		t.Errorf("the opponent who discarded lost %d, want 5", got)
	}
	if no.Life != noLife {
		t.Errorf("the opponent who declined lost %d", noLife-no.Life)
	}
	if me.Hand.Size() != 5 || yes.Hand.Size() != 5 {
		t.Errorf("hands after the wheel: me %d, opponent %d, want 5 and 5", me.Hand.Size(), yes.Hand.Size())
	}
}

func TestDeck3655AngelExilesUpToTwoArtifactsAndEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := pushCatalogPermanent(g, me.ID, "Angel of the Ruins", "Artifact Creature — Angel", deck3655AngelOracle, false)
	a := pushCostedPermanentForTest(g, opp.ID, "Rock", "Artifact", "{1}")
	b := pushCostedPermanentForTest(g, opp.ID, "Aura", "Enchantment", "{1}")
	spec, _ := Lookup(deck3655AngelOracle)
	item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: angel,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1}}}
	if err := spec.Triggered[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a targeted artifact or enchantment was not exiled")
	}
	if len(spec.Activated) != 1 {
		t.Errorf("activated abilities = %d, want plainscycling only", len(spec.Activated))
	}
}

func TestDeck3655ShapesOfTheSimpleCards(t *testing.T) {
	grinder, _ := Lookup(deck3655GrinderOracle)
	if len(grinder.Triggered) != 1 || len(grinder.Activated) != 1 {
		t.Errorf("Ruin Grinder: %d triggers, %d activated, want 1 and 1", len(grinder.Triggered), len(grinder.Activated))
	}
	phoenix, _ := Lookup(deck3655PhoenixOracle)
	if len(phoenix.Activated) != 1 || phoenix.Activated[0].Modes == nil {
		t.Error("Phoenix Down must be one modal activated ability")
	}
	key, _ := Lookup(deck3655KeyOracle)
	if len(key.Activated) != 1 || len(key.Triggered) != 1 {
		t.Errorf("Key to the City: %d activated, %d triggers, want 1 and 1", len(key.Activated), len(key.Triggered))
	}
}
