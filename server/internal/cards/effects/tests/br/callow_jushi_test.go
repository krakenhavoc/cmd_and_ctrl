package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsbu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func callowJushiRow() cards.Card {
	return effectstest.FlipRow(cardsbu.CallowJushiOracleID, "{1}{U}{U}", []string{"U"}, nil,
		cards.CardFace{Name: "Callow Jushi", TypeLine: "Creature — Human Wizard",
			OracleText: "Whenever you cast a Spirit or Arcane spell, you may put a ki counter on this creature.\nAt the beginning of the end step, if there are two or more ki counters on this creature, you may flip it.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Jaraku the Interloper", TypeLine: "Legendary Creature — Spirit",
			OracleText: "Remove a ki counter from Jaraku: Counter target spell unless its controller pays {2}.",
			Power:      "3", Toughness: "4"})
}

func TestCallowJushiKiCountersAndFlip(t *testing.T) {
	effectstest.KiCycleCheck(t, callowJushiRow, cardsbu.CallowJushiOracleID, "Callow Jushi", "Jaraku the Interloper", []string{"U"}, 3)
}

// Jaraku counters a spell unless its controller pays {2}: the prompt
// goes to the spell's controller, who may pay or let it be countered,
// and the stack waits for the answer.
func TestJarakuCountersUnlessItsControllerPays(t *testing.T) {
	for _, tc := range []struct {
		name string
		pay  bool
	}{{"pays", true}, {"declines", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			effectstest.ToMainForCost(t, g)
			jaraku := effectstest.FlipEntered(t, g, callowJushiRow(), them)
			effectstest.PfAddCounters(g, jaraku, "ki", 2)
			g.WithWriteLock(func() {
				if err := g.FlipPermanentForEffect(jaraku); err != nil {
					t.Fatal(err)
				}
			})
			victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority to the responder: %v", err)
			}
			if err := g.ActivateCatalogAbility(them.ID, jaraku, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err != nil {
				t.Fatalf("activate Jaraku: %v", err)
			}
			if n := effectstest.KiCount(t, g, jaraku); n != 1 {
				t.Errorf("%d ki counters, want 1 (one removed as the cost)", n)
			}
			effectstest.PassUntilTaxed(t, g, me.ID)
			if !g.Stack.Contains(victim) {
				t.Fatal("the spell resolved while the {2} was unanswered")
			}
			if tc.pay {
				effectstest.FloatMana(t, g, me, "{C}{C}")
			}
			effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
			effectstest.PassPriorityAroundTable(t, g)
			if countered := effectstest.P7Countered(g, victim); countered == tc.pay {
				t.Errorf("pay = %v but countered = %v", tc.pay, countered)
			}
		})
	}
}
