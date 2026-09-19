package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_provenance_cards_test.go — the CARD half of #653 (CR 400.7d).
// The engine half, against fixtures with no catalog entry, is
// server/internal/game/cast_provenance_test.go.
//
// One question, asked of three cards: does the permanent remember that
// the spell which became it was cast for its escape cost? Phlage and
// Uro sacrifice themselves when it did not; Pharika's Spawn does
// nothing extra unless it did.

const (
	phlageOracle        = "3407eb6e-b74d-4159-a801-d7163937953c"
	uroOracle           = "ee302659-59ed-4eef-babe-451b9ccf7f14"
	pharikasSpawnOracle = "a44955e7-f1ad-41b9-b93d-0a980ac341d8"
)

// escapeIntoPlay casts a catalog creature out of the active seat's
// graveyard for its escape cost, paying with `exile` fresh fodder
// cards, and returns the card's ID.
func escapeIntoPlay(t *testing.T, g *game.Game, name, typeLine, oracle string, exile int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, name, typeLine, oracle)
	pay := seedGraveyardFodder(t, g, exile)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast of %s: %v", name, err)
	}
	return id
}

// handCastIntoPlay casts the same catalog creature from hand for its
// printed cost.
func handCastIntoPlay(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, name, typeLine, oracle, nil)
}

// Phlage escaped: the 6/6 stays, the Helix still happens.
func TestPhlageThatEscapedIsNotSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	life := opp.Life
	mine := me.Life

	id := escapeIntoPlay(t, g, "Phlage, Titan of Fire's Fury", "Creature — Elder Giant", phlageOracle, 5)
	passPriorityAroundTable(t, g)
	// The damage trigger targets, so it asks as it goes on the stack.
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("an escaped Phlage sacrificed itself")
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want 3 damage", life, opp.Life)
	}
	if me.Life != mine+3 {
		t.Errorf("my life %d → %d, want 3 gained", mine, me.Life)
	}
}

// Phlage hard-cast: the Helix happens and the 6/6 does not stay. The
// whole card, and the reason the permanent has to remember.
func TestPhlageHardCastSacrificesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	life := opp.Life

	id := handCastIntoPlay(t, g, "Phlage, Titan of Fire's Fury", "Creature — Elder Giant", phlageOracle)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(id) {
		t.Error("a hard-cast Phlage stayed on the battlefield")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the sacrificed Phlage is not in its owner's graveyard")
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d — the Helix happens either way", life, opp.Life)
	}
}

// And the CR 400.7 half on a real card: a Phlage that escaped, died
// and was reanimated is a NEW object that did not escape, so the
// second entry sacrifices it. Without the clear in MoveCard this is
// the infinite loop the card is famous for not being.
func TestAReanimatedPhlageDoesNotRememberEscaping(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	id := escapeIntoPlay(t, g, "Phlage, Titan of Fire's Fury", "Creature — Elder Giant", phlageOracle, 5)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, g.Seats[1].ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the escaped Phlage never landed")
	}

	// Kill it and put it straight back, the way a reanimation spell
	// would — no cast, so no provenance.
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Battlefield, me.Graveyard, id); err != nil {
			t.Fatalf("to the graveyard: %v", err)
		}
		if _, err := game.MoveCard(me.Graveyard, g.Battlefield, id); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	if g.CastProvenanceForEffect(id).Escaped() {
		t.Error("a reanimated Phlage still reports escaped")
	}
}

// Uro, the same shape with the other rider: the land is offered, and
// an escaped Uro is not sacrificed.
func TestUroThatEscapedStaysAndOffersTheLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()

	id := escapeIntoPlay(t, g, "Uro, Titan of Nature's Wrath", "Creature — Elder Giant", uroOracle, 5)
	passPriorityAroundTable(t, g)
	// CR 603.3b: both of Uro's triggers are its controller's, so they
	// go on the stack in an order the controller picks.
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("an escaped Uro sacrificed itself")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want 3 gained", life, me.Life)
	}
	// One card drawn. The hand may also be holding a land prompt open,
	// which is a choice and not a card movement.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestUroHardCastSacrificesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	id := handCastIntoPlay(t, g, "Uro, Titan of Nature's Wrath", "Creature — Elder Giant", uroOracle)
	passPriorityAroundTable(t, g)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(id) {
		t.Error("a hard-cast Uro stayed on the battlefield")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the sacrificed Uro is not in its owner's graveyard")
	}
}

// Pharika's Spawn is "when it enters THIS WAY" — the trigger exists
// either way and does nothing unless the permanent escaped.
func TestPharikasSpawnOnlyEatsACreatureWhenItEscaped(t *testing.T) {
	for _, escaped := range []bool{true, false} {
		name := "hard-cast"
		if escaped {
			name = "escaped"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			// A creature for the opponent to lose, and a Gorgon they
			// keep whatever happens.
			bear := uuid.New()
			gorgon := uuid.New()
			g.WithWriteLock(func() {
				g.Battlefield.PushTop(game.Card{
					InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
					Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
				})
				g.Battlefield.PushTop(game.Card{
					InstanceID: gorgon, Name: "Spare Gorgon", TypeLine: "Creature — Gorgon",
					Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
				})
			})

			var id uuid.UUID
			if escaped {
				id = escapeIntoPlay(t, g, "Pharika's Spawn", "Creature — Gorgon", pharikasSpawnOracle, 3)
			} else {
				id = handCastIntoPlay(t, g, "Pharika's Spawn", "Creature — Gorgon", pharikasSpawnOracle)
			}
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatal("Pharika's Spawn never landed")
			}

			pick := sacrificeChoiceFor(g, opp.ID)
			if !escaped {
				if pick != nil {
					t.Fatal("a hard-cast Pharika's Spawn asked an opponent to sacrifice")
				}
				return
			}
			if pick == nil {
				t.Fatal("an escaped Pharika's Spawn asked nobody to sacrifice")
			}
			answerSacrifice(t, g, opp.ID, bear)
			if g.Battlefield.Contains(bear) {
				t.Error("the chosen creature was not sacrificed")
			}
			if !g.Battlefield.Contains(gorgon) {
				t.Error("the Gorgon was sacrificed — the clause says non-Gorgon")
			}
			// CR 702.138c: and it arrived as a 5/6.
			c, ok := g.LookupCardForEffect(id)
			if !ok || c.Counters[game.CounterPlusOne] != 2 {
				t.Errorf("escape counters = %d, want 2", c.Counters[game.CounterPlusOne])
			}
		})
	}
}
