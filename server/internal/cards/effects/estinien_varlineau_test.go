package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// estinien_varlineau_test.go — #3786: "opponents dealt combat damage by
// Estinien Varlineau or a Dragon this turn", and the noncreature-spell pump.

const estinienOracle = "29cf1dc2-3a6a-4a8b-8d61-4b6f0e74f271"

func estinienTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Estinien Varlineau", "Legendary Creature — Elf Warrior", estinienOracle, false)
	return g, me, id
}

func TestEstinienPumpsAndFliesOnANoncreatureSpellOnly(t *testing.T) {
	g, _, id := estinienTable(t)
	if hasEffectiveKeyword(t, g, id, "flying") {
		t.Fatal("setup: Estinien has no printed flying")
	}
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature spell put %d counters on Estinien", got)
	}
	b24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("a noncreature spell: %d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Estinien did not gain flying")
	}
}

func TestEstinienDrawsAndLosesForDistinctOpponentsHitByHimOrAnyDragon(t *testing.T) {
	g, me, id := estinienTable(t)
	opp1, opp2, opp3 := g.Seats[1], g.Seats[2], g.Seats[3]
	// Opponent one's Dragon hits opponent two: "a Dragon", not "a Dragon
	// you control".
	dragon := b12Creature(g, opp1.ID, "Opposing Dragon", "Creature — Dragon", 4, 4)
	dealCombatDamage(g, opp1.ID, dragon, opp2.ID)
	// Estinien hits opponent one twice; the opponent counts once.
	dealCombatDamage(g, me.ID, id, opp1.ID)
	dealCombatDamage(g, me.ID, id, opp1.ID)
	// A Dragon hitting the controller is not an opponent, and a Bear
	// hitting opponent three is not a Dragon.
	dealCombatDamage(g, opp1.ID, dragon, me.ID)
	bear := b12Creature(g, opp2.ID, "Bear", "Creature — Bear", 2, 2)
	dealCombatDamage(g, opp2.ID, bear, opp3.ID)

	hand, life := handSizeOf(g, me), me.Life
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me) - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (opponents one and two)", got)
	}
	if got := life - me.Life; got != 2 {
		t.Errorf("lost %d life, want 2", got)
	}
}

func TestEstinienCountsAChangelingAsADragon(t *testing.T) {
	g, me, _ := estinienTable(t)
	opp := g.Seats[1]
	shifter := pushCatalogPermanent(g, opp.ID, "Shifter", "Creature — Shapeshifter", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == shifter {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, game.KeywordChangeling)
			}
		}
	})
	dealCombatDamage(g, opp.ID, shifter, g.Seats[2].ID)
	hand := handSizeOf(g, me)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me) - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (a changeling is a Dragon)", got)
	}
}

func TestEstinienWithNobodyHitDoesNothing(t *testing.T) {
	g, me, _ := estinienTable(t)
	hand, life := handSizeOf(g, me), me.Life
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if handSizeOf(g, me) != hand || me.Life != life {
		t.Errorf("X of zero changed hand %d->%d or life %d->%d", hand, handSizeOf(g, me), life, me.Life)
	}
}

// Estinien dies with the ability on the stack: the object still dealt
// the damage (CR 608.2h), so the opponent still counts.
func TestEstinienSacrificedInResponseStillCountsHisOwnDamage(t *testing.T) {
	g, me, id := estinienTable(t)
	opp := g.Seats[1]
	dealCombatDamage(g, me.ID, id, opp.ID)
	hand := handSizeOf(g, me)
	advanceTo(t, g, game.StepPostcombatMain)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me) - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}
