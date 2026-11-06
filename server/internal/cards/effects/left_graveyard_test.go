package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// left_graveyard_test.go — #4899: "whenever one or more cards leave your
// graveyard" counts every way a card leaves it. A flashback cast (an
// EventCast from the graveyard, with no zone-move event of its own), a
// land played from it (Crucible of Worlds), an exile from it. A card
// leaving an opponent's graveyard does not count, and "one or more" is
// still one trigger for cards leaving as one event.
//
// Every case drives the engine: CastSpell, the land play, and the exile
// effect, with the catalog card on the battlefield.

const crucibleOfWorldsOracle = "33c722cf-b4bf-431f-aefd-ee96241a7fbf"

// leftGraveyardCard is one catalog card with a "leaves your graveyard"
// trigger. creatureOnly is "creature card(s)"; perCard is a trigger
// without "one or more" (one per card, not one per batch).
type leftGraveyardCard struct {
	name, typeLine, oracle string
	creatureOnly, perCard  bool
}

var leftGraveyardCards = []leftGraveyardCard{
	{name: "Quintorius, Field Historian", typeLine: "Legendary Creature — Elephant Cleric", oracle: socB11FieldHistorianOracle},
	{name: "Quintorius, History Chaser", typeLine: "Legendary Planeswalker — Quintorius", oracle: socB11HistoryChaserOracle},
	{name: "Spirit of Resilience", typeLine: "Creature — Spirit Warrior", oracle: "ddc85b23-113a-46e9-ab09-fae9312c72fa"},
	{name: "Tormod, the Desecrator", typeLine: "Legendary Creature — Zombie Wizard", oracle: "6dc0150d-7145-41e2-bd3d-2564d9d32301"},
	{name: "Teval, the Balanced Scale", typeLine: "Legendary Creature — Spirit Dragon", oracle: "c8cbf0ec-ec98-4cb3-8068-60e92bbd740d"},
	{name: "On Wings of Gold", typeLine: "Enchantment", oracle: "3b6f608d-f27d-44cd-b0a1-0e1e1aa1c98a"},
	{name: "Ark of Hunger", typeLine: "Artifact", oracle: socB20ArkOracle},
	{name: "Desecrated Tomb", typeLine: "Artifact", oracle: "07ff619a-21ee-44b6-b666-ceab4a78096a", creatureOnly: true},
	{name: "Grim Reaper's Scythe", typeLine: "Legendary Artifact — Equipment", oracle: "91b1db8c-4bce-43d1-94ab-14b1435686c7", creatureOnly: true},
	{name: "Insidious Roots", typeLine: "Enchantment", oracle: "d75b2b8e-05c9-47da-b359-a867256d78ea", creatureOnly: true},
	{name: "Cyan, Vengeful Samurai", typeLine: "Legendary Creature — Human Samurai", oracle: "5a2fa420-c27f-4c14-b285-e0bb2828e053", creatureOnly: true},
	{name: "Undead Hand Ninja", typeLine: "Creature — Zombie Ninja", oracle: "3c02d725-a05e-4086-9caa-ddb005d7ae18", creatureOnly: true},
	{name: "Along the Crooked Way", typeLine: "Enchantment", oracle: "f4437cb3-85f7-49b5-bb41-6a948549d987", creatureOnly: true, perCard: true},
	{name: "Syr Konrad, the Grim", typeLine: "Legendary Creature — Human Knight", oracle: "14c3ff84-1e82-4606-a433-869fc52cc382", creatureOnly: true, perCard: true},
}

// leftGraveyardTable is a fresh game at seat 0's main phase with the card
// on seat 0's battlefield.
func leftGraveyardTable(t *testing.T, tc leftGraveyardCard) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMainOf(t, g, 0)
	me, opp := g.Seats[0], g.Seats[1]
	var src uuid.UUID
	if tc.typeLine == "Legendary Planeswalker — Quintorius" {
		src = socB11Chaser(g, me, 5)
	} else {
		src = b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
	}
	return g, me, opp, src
}

// leftGraveyardTriggers counts src's triggered items, on the stack or
// still waiting to be put there (an effect run straight from a test
// leaves its triggers pending until the next state check).
func leftGraveyardTriggers(g *game.Game, src uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == src {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == src {
			n++
		}
	}
	return n
}

// leftGraveyardCastGravecrawler casts Gravecrawler (a creature card that
// may be cast from its owner's graveyard while they control a Zombie)
// from me's graveyard.
func leftGraveyardCastGravecrawler(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	b12Push(g, me.ID, "My Zombie", "Creature — Zombie", "", 2, 2)
	id := condZoneGraveyardCard(t, g, me, "Gravecrawler", "Creature — Zombie", "{B}", gravecrawlerOracle)
	if err := condZoneCastFromGraveyard(g, me, id); err != nil {
		t.Fatalf("cast Gravecrawler from the graveyard: %v", err)
	}
	return id
}

func TestLeftGraveyardFlashbackCastTriggers(t *testing.T) {
	for _, tc := range leftGraveyardCards {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, src := leftGraveyardTable(t, tc)
			looting := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
			if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
				t.Fatalf("flashback: %v", err)
			}
			want := 1
			if tc.creatureOnly {
				want = 0 // a sorcery card is not a creature card
			}
			if got := leftGraveyardTriggers(g, src); got != want {
				t.Errorf("%d triggers for a flashback cast of a sorcery from your graveyard, want %d", got, want)
			}
		})
	}
}

func TestLeftGraveyardCreatureCastFromTheGraveyardTriggers(t *testing.T) {
	for _, tc := range leftGraveyardCards {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, src := leftGraveyardTable(t, tc)
			leftGraveyardCastGravecrawler(t, g, me)
			if got := leftGraveyardTriggers(g, src); got != 1 {
				t.Errorf("%d triggers for a creature card cast from your graveyard, want 1", got)
			}
		})
	}
}

func TestLeftGraveyardLandPlayedFromTheGraveyardTriggers(t *testing.T) {
	for _, tc := range leftGraveyardCards {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, src := leftGraveyardTable(t, tc)
			pushCatalogPermanent(g, me.ID, "Crucible of Worlds", "Artifact", crucibleOfWorldsOracle, false)
			land := pushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
			if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
				t.Fatalf("play a land from the graveyard: %v", err)
			}
			if !onBattlefield(g, land) {
				t.Fatal("the land is not on the battlefield")
			}
			want := 1
			if tc.creatureOnly {
				want = 0 // a land card is not a creature card
			}
			if got := leftGraveyardTriggers(g, src); got != want {
				t.Errorf("%d triggers for a land played from your graveyard, want %d", got, want)
			}
		})
	}
}

func TestLeftGraveyardExileTriggersAndAnOpponentsGraveyardDoesNot(t *testing.T) {
	for _, tc := range leftGraveyardCards {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, src := leftGraveyardTable(t, tc)
			theirs := socB11GraveyardCreature(opp, "Their Dead Bear")
			g.WithWriteLock(func() {
				if err := g.ExileCardForEffect(theirs); err != nil {
					t.Fatalf("exile: %v", err)
				}
			})
			if got := leftGraveyardTriggers(g, src); got != 0 {
				t.Fatalf("%d triggers for a card leaving an opponent's graveyard, want 0", got)
			}
			mine := socB11GraveyardCreature(me, "My Dead Bear")
			g.WithWriteLock(func() {
				if err := g.ExileCardForEffect(mine); err != nil {
					t.Fatalf("exile: %v", err)
				}
			})
			if got := leftGraveyardTriggers(g, src); got != 1 {
				t.Errorf("%d triggers for a card exiled from your graveyard, want 1", got)
			}
		})
	}
}

func TestLeftGraveyardTwoCardsExiledAtOnceKeepBatchSemantics(t *testing.T) {
	for _, tc := range leftGraveyardCards {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, src := leftGraveyardTable(t, tc)
			a := socB11GraveyardCreature(me, "Dead One")
			b := socB11GraveyardCreature(me, "Dead Two")
			g.WithWriteLock(func() {
				if err := g.ExileCardsKeyedForEffect([]uuid.UUID{a, b}, game.KeyedTail{}); err != nil {
					t.Fatalf("exile: %v", err)
				}
			})
			want := 1 // "one or more": one trigger for one event
			if tc.perCard {
				want = 2
			}
			if got := leftGraveyardTriggers(g, src); got != want {
				t.Errorf("%d triggers for two cards exiled at once, want %d", got, want)
			}
		})
	}
}

// Relic Retriever and Gau read "a card left your graveyard this turn" at
// the end step: a flashback cast this turn counts.
func TestLeftGraveyardThisTurnCountsAFlashbackCast(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Relic Retriever", "Creature — Spirit Monkey", socB11RetrieverOracle},
		{"Gau, Feral Youth", "Legendary Creature — Human Berserker", "9e96d12b-8381-4706-8089-3521aa2b5422"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMainOf(t, g, 0)
			me := g.Seats[0]
			src := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			looting := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
			if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
				t.Fatalf("flashback: %v", err)
			}
			if err := g.CounterSpell(looting, nil); err != nil {
				t.Fatalf("counter: %v", err)
			}
			advanceToStepOf(t, g, 0, game.StepEnd)
			if triggerOnStack(g, src) == nil {
				t.Error("no end-step trigger after a flashback cast from your graveyard this turn")
			}
		})
	}
}

// Spirit of Resilience's "from among those cards" includes a creature card
// cast from the graveyard: the copy choice offers it.
func TestLeftGraveyardSpiritOfResilienceOffersTheCastCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMainOf(t, g, 0)
	me := g.Seats[0]
	spirit := b12Push(g, me.ID, "Spirit of Resilience", "Creature — Spirit Warrior", "ddc85b23-113a-46e9-ab09-fae9312c72fa", 2, 2)
	crawler := leftGraveyardCastGravecrawler(t, g, me)
	if triggerOnStack(g, spirit) == nil {
		t.Fatal("no trigger for a creature card cast from your graveyard")
	}
	// The trigger is on top of the Gravecrawler spell: one round of passes
	// resolves it alone.
	passPriorityAroundTable(t, g)
	c := pendingChooseCards(t, g, me.ID)
	found := false
	for _, id := range c.ChooseCards {
		found = found || id == crawler
	}
	if !found {
		t.Errorf("the copy choice %v does not offer Gravecrawler, the card cast from your graveyard", c.ChooseCards)
	}
}

// #4943: two departures made by two separate actions are two events, even
// with nothing resolving between them. Casting a spell (CR 601.2a moves the
// card to the stack) and playing a land (CR 116.2a) are each their own
// event, so a "one or more" trigger triggers once for each (CR 603.2c),
// not once for the priority window both happened in.

func TestLeftGraveyardTwoFlashbackCastsInOneWindowTriggerTwice(t *testing.T) {
	for _, tc := range leftGraveyardCards {
		if tc.creatureOnly {
			continue // neither card is a creature card
		}
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, src := leftGraveyardTable(t, tc)
			looting := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
			if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
				t.Fatalf("flashback Faithless Looting: %v", err)
			}
			// An instant, so it can be cast in the same window, on top of the
			// Looting and the trigger it caused, with nothing resolved.
			iteration := seedGraveyardCard(t, g, "Galvanic Iteration", "Instant", galvanicIterationOracle)
			if err := g.CastSpell(me.ID, iteration, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
				t.Fatalf("flashback Galvanic Iteration: %v", err)
			}
			if got := leftGraveyardTriggers(g, src); got != 2 {
				t.Errorf("%d triggers for two flashback casts in one priority window, want 2", got)
			}
		})
	}
}

func TestLeftGraveyardALandPlayThenACastInOneWindowTriggerTwice(t *testing.T) {
	for _, tc := range leftGraveyardCards {
		if tc.creatureOnly {
			continue // neither card is a creature card
		}
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, src := leftGraveyardTable(t, tc)
			pushCatalogPermanent(g, me.ID, "Crucible of Worlds", "Artifact", crucibleOfWorldsOracle, false)
			land := pushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
			if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
				t.Fatalf("play a land from the graveyard: %v", err)
			}
			iteration := seedGraveyardCard(t, g, "Galvanic Iteration", "Instant", galvanicIterationOracle)
			if err := g.CastSpell(me.ID, iteration, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
				t.Fatalf("flashback Galvanic Iteration: %v", err)
			}
			if got := leftGraveyardTriggers(g, src); got != 2 {
				t.Errorf("%d triggers for a land play and a flashback cast in one priority window, want 2", got)
			}
		})
	}
}
