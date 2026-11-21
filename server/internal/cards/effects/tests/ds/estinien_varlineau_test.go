package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// estinien_varlineau_test.go — #3786: "opponents dealt combat damage by
// Estinien Varlineau or a Dragon this turn", and the noncreature-spell pump.

const estinienOracle = "29cf1dc2-3a6a-4a8b-8d61-4b6f0e74f271"

func estinienTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Estinien Varlineau", "Legendary Creature — Elf Warrior", estinienOracle, false)
	return g, me, id
}

func TestEstinienPumpsAndFliesOnANoncreatureSpellOnly(t *testing.T) {
	g, _, id := estinienTable(t)
	if effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Fatal("setup: Estinien has no printed flying")
	}
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature spell put %d counters on Estinien", got)
	}
	effectstest.B24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("a noncreature spell: %d +1/+1 counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Estinien did not gain flying")
	}
}

func TestEstinienDrawsAndLosesForDistinctOpponentsHitByHimOrAnyDragon(t *testing.T) {
	g, me, id := estinienTable(t)
	opp1, opp2, opp3 := g.Seats[1], g.Seats[2], g.Seats[3]
	// Opponent one's Dragon hits opponent two: "a Dragon", not "a Dragon
	// you control".
	dragon := effectstest.B12Creature(g, opp1.ID, "Opposing Dragon", "Creature — Dragon", 4, 4)
	effectstest.DealCombatDamage(g, opp1.ID, dragon, opp2.ID)
	// Estinien hits opponent one twice; the opponent counts once.
	effectstest.DealCombatDamage(g, me.ID, id, opp1.ID)
	effectstest.DealCombatDamage(g, me.ID, id, opp1.ID)
	// A Dragon hitting the controller is not an opponent, and a Bear
	// hitting opponent three is not a Dragon.
	effectstest.DealCombatDamage(g, opp1.ID, dragon, me.ID)
	bear := effectstest.B12Creature(g, opp2.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamage(g, opp2.ID, bear, opp3.ID)

	hand, life := effectstest.HandSizeOf(g, me), me.Life
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me) - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (opponents one and two)", got)
	}
	if got := life - me.Life; got != 2 {
		t.Errorf("lost %d life, want 2", got)
	}
}

func TestEstinienCountsAChangelingAsADragon(t *testing.T) {
	g, me, _ := estinienTable(t)
	opp := g.Seats[1]
	shifter := effectstest.PushCatalogPermanent(g, opp.ID, "Shifter", "Creature — Shapeshifter", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == shifter {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, game.KeywordChangeling)
			}
		}
	})
	effectstest.DealCombatDamage(g, opp.ID, shifter, g.Seats[2].ID)
	hand := effectstest.HandSizeOf(g, me)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me) - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (a changeling is a Dragon)", got)
	}
}

func TestEstinienWithNobodyHitDoesNothing(t *testing.T) {
	g, me, _ := estinienTable(t)
	hand, life := effectstest.HandSizeOf(g, me), me.Life
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSizeOf(g, me) != hand || me.Life != life {
		t.Errorf("X of zero changed hand %d->%d or life %d->%d", hand, effectstest.HandSizeOf(g, me), life, me.Life)
	}
}

// Estinien dies with the ability on the stack: the object still dealt
// the damage (CR 608.2h), so the opponent still counts.
func TestEstinienSacrificedInResponseStillCountsHisOwnDamage(t *testing.T) {
	g, me, id := estinienTable(t)
	opp := g.Seats[1]
	effectstest.DealCombatDamage(g, me.ID, id, opp.ID)
	hand := effectstest.HandSizeOf(g, me)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me) - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}
