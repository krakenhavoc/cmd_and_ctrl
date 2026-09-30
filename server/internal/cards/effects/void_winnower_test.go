package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const voidWinnowerOracle = "70ac902e-1eb5-4f83-a5a6-00ac89cfe50f"

// TestVoidWinnowerStopsOpponentsCastingEvenManaValueSpells — an
// opponent's even-MV spell is refused; an odd-MV spell is not.
func TestVoidWinnowerStopsOpponentsCastingEvenManaValueSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Void Winnower", "Creature — Eldrazi", voidWinnowerOracle, false)

	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	evenID := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: evenID, Name: "Two Drop", TypeLine: "Sorcery", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, evenID, game.CastSpellParams{}); err == nil {
		t.Error("an opponent casting an even mana value spell should be refused")
	}

	oddID := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: oddID, Name: "One Drop", TypeLine: "Sorcery", ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, oddID, game.CastSpellParams{}); err != nil {
		t.Errorf("an opponent casting an odd mana value spell should be allowed: %v", err)
	}
}

// TestVoidWinnowerStopsOpponentsBlockingWithEvenManaValueCreatures —
// checked at the pair-legality level, the same gate block declaration
// uses.
func TestVoidWinnowerStopsOpponentsBlockingWithEvenManaValueCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Void Winnower", "Creature — Eldrazi", voidWinnowerOracle, false)

	attacker := seedCreature(g, "Attacker", me.ID)
	evenBlocker := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: evenBlocker, Name: "Even Blocker", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	oddBlocker := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: oddBlocker, Name: "Odd Blocker", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})

	var canBlockEven, canBlockOdd bool
	g.ReadSnapshot(func() {
		var att, evenC, oddC *game.Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case attacker:
				att = c
			case evenBlocker:
				evenC = c
			case oddBlocker:
				oddC = c
			}
		}
		canBlockEven = g.CanBlockLocked(att, evenC)
		canBlockOdd = g.CanBlockLocked(att, oddC)
	})
	if canBlockEven {
		t.Error("an opponent's even mana value creature should not be able to block")
	}
	if !canBlockOdd {
		t.Error("an opponent's odd mana value creature should still be able to block")
	}
}
