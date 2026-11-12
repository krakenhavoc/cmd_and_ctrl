package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cant_pay_life_test.go — #5660. Under Karn's Sylex ("Players can't pay
// life to cast spells or to activate abilities that aren't mana
// abilities") the engine refuses every life component of a cast and of a
// non-mana activation, so the enumerator must offer none: a move it
// offered would be one the engine refuses (#544).

const oracleKarnsSylex = "c5e817f9-a003-45b5-8e16-1746e6096166"

// sylexOn seats the real catalog Karn's Sylex under `p`. Its rule is a
// static read off the layered view, so the entry is announced and the
// layers caught up, as they are after any real entry.
func sylexOn(g *game.Game, p *game.Player) uuid.UUID {
	id := battlefieldCard(g, p, game.Card{
		Name: "Karn's Sylex", TypeLine: "Legendary Artifact", ManaCost: "{3}", OracleID: oracleKarnsSylex,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.RunStateChecksForTest()
	return id
}

// Dismember with one Swamp is offered for 4 life; with any player's
// Sylex out it is not offered at all, and with three Swamps only the
// mana payment is.
func TestKarnsSylexOffersNoPhyrexianLifeCast(t *testing.T) {
	g, seat, card := phyrexianDismember(t, 1, 20)
	if got := lifeCounts(t, legal.EnumerateFor(g, seat.ID), card); !sameInts(got, []int{2}) {
		t.Fatalf("without the Sylex the life counts are %v, want [2]", got)
	}
	sylexOn(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)])
	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), card); len(got) != 0 {
		t.Fatalf("under Karn's Sylex a cast that needs life is offered: %v", labels(got))
	}

	g, seat, card = phyrexianDismember(t, 3, 20)
	sylexOn(g, seat)
	moves := legal.EnumerateFor(g, seat.ID)
	if got := lifeCounts(t, moves, card); !sameInts(got, []int{0}) {
		t.Fatalf("under Karn's Sylex with the mana the life counts are %v, want [0]", got)
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, card))
}

// A non-mana activation is offered neither with a printed life cost nor
// with a Phyrexian symbol bought with life.
func TestKarnsSylexOffersNoLifePaidActivation(t *testing.T) {
	for _, withSylex := range []bool{false, true} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		mark := func(_ *game.Game, _ *game.StackItem) error { return nil }
		printed := battlefieldCard(g, seat, game.Card{
			Name: "Life Engine", TypeLine: "Artifact",
			ActivatedAbilities: []game.ActivatedAbilityShape{{
				Label: "Pay 2 life: Mark", Cost: game.AbilityCost{Life: 2}, Effect: mark,
			}},
		})
		phyrexian := battlefieldCard(g, seat, game.Card{
			Name: "Phyrexian Engine", TypeLine: "Artifact",
			ActivatedAbilities: []game.ActivatedAbilityShape{{
				Label: "{B/P}: Mark", Cost: game.AbilityCost{Mana: "{B/P}"}, Effect: mark,
			}},
		})
		if withSylex {
			sylexOn(g, seat)
		}
		advanceTo(t, g, game.StepPrecombatMain)
		moves := legal.EnumerateFor(g, seat.ID)
		for name, id := range map[string]uuid.UUID{"printed life cost": printed, "Phyrexian symbol": phyrexian} {
			got := activationsOf(moves, id)
			switch {
			case withSylex && len(got) != 0:
				t.Errorf("under Karn's Sylex the %s activation is offered: %v", name, labels(got))
			case !withSylex && len(got) != 1:
				t.Errorf("without the Sylex the %s activation is offered %d times, want 1: %v", name, len(got), labels(moves))
			}
		}
	}
}
